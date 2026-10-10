"""cdecls.py DECOMP: every global's C type as each translation unit of the decompilation declares it (clang AST),
printed as JSON {name: {"types": {type: [files]}, "defined": file}}. far pointers keep a "far" marker."""
import json, os, subprocess, sys, glob
src = os.path.join(sys.argv[1], 'src')
flags = [l.strip() for l in open(os.path.join(src, 'compile_flags.txt')) if l.strip()]
flags = [f for f in flags if f not in ('-include', '../tools/clangd/bcc.h')]
shim = os.path.join(os.environ.get('TMPDIR', '/tmp'), 'far_shim.h')
open(shim, 'w').write('#include "%s"\n#undef far\n#define far __attribute__((address_space(1)))\n'
                      % os.path.join(sys.argv[1], 'tools/clangd/bcc.h'))
out = {}
for c in sorted(glob.glob(os.path.join(src, '*.c'))):
    r = subprocess.run(['clang', '-fsyntax-only', '-fpack-struct=1'] + flags + ['-include', shim, '-Xclang', '-ast-dump=json', os.path.basename(c)],
                       cwd=src, capture_output=True, text=True)
    ast = json.loads(r.stdout)
    for d in ast.get('inner', []):
        if d.get('kind') != 'VarDecl' or 'name' not in d: continue
        t = d['type'].get('desugaredQualType', d['type']['qualType']).replace('__attribute__((address_space(1)))', 'far')
        e = out.setdefault(d['name'], {'types': {}, 'defined': None})
        e['types'].setdefault(t, []).append(os.path.basename(c))
        if d.get('storageClass') != 'extern': e['defined'] = os.path.basename(c)
json.dump(out, sys.stdout, indent=1, sort_keys=True)
