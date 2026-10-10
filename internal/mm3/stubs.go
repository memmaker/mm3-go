package mm3

// Stubs for the callees of the ported modules that are not ported yet, with their final
// signatures. Whoever ports the real one deletes its stub.

func classCanUseItem(cls, id, quiet int) int                     { panic("todo: classCanUseItem") }       // stub: shop.c
func calcMaxHp(c *Chr) int                                       { panic("todo: calcMaxHp") }             // stub: shop.c
func charSavingThrow(c *Chr, kind int) int                       { panic("todo: charSavingThrow") }       // stub: combat.c
func monsterSavingThrow(lvl int) int                             { panic("todo: monsterSavingThrow") }    // stub: combat.c
func monsterTurn(mt int)                                         { panic("todo: monsterTurn") }           // stub: combat.c
func drawCompassButton(k int)                                    { panic("todo: drawCompassButton") }     // stub: items.c
func showPleaseWait()                                            { panic("todo: showPleaseWait") }        // stub: items.c
func bounceDice()                                                { panic("todo: bounceDice") }            // stub: guild.c
func randomRange(a, b int) int                                   { panic("todo: randomRange") }           // stub: sys.asm
func drawMazeView()                                              { panic("todo: drawMazeView") }          // stub: itemname.c
func runCombat()                                                 { panic("todo: runCombat") }             // stub: combat.c
func spellAttackMonsters(n int)                                  { panic("todo: spellAttackMonsters") }   // stub: items.c
func getThievery(c *Chr) int                                     { panic("todo: getThievery") }           // stub: items.c
func playEnding(score uint32) int                                { panic("todo: playEnding") }            // stub: ending.c
func faceYesNoDialog(face int, a, b []byte, mode, shown int) int { panic("todo: faceYesNoDialog") }       // stub: building.c
func visitBank()                                                 { panic("todo: visitBank") }             // stub: building.c
func visitSmithy()                                               { panic("todo: visitSmithy") }           // stub: building.c
func visitGuild()                                                { panic("todo: visitGuild") }            // stub: building.c
func visitInn() int                                              { panic("todo: visitInn") }              // stub: building.c
func visitTavern()                                               { panic("todo: visitTavern") }           // stub: building.c
func visitTemple()                                               { panic("todo: visitTemple") }           // stub: building.c
func visitTraining()                                             { panic("todo: visitTraining") }         // stub: building.c
func buildingWaitInput() int                                     { panic("todo: buildingWaitInput") }     // stub: building.c
func calcArmorClass(c *Chr, base int) int                        { panic("todo: calcArmorClass") }        // stub: shop.c
func calcMaxSp(c *Chr) int                                       { panic("todo: calcMaxSp") }             // stub: shop.c
func buildingEntryDialog(title, seq []byte, n, typ int) int      { panic("todo: buildingEntryDialog") }   // stub: building.c
func showCoraksNotes()                                           { panic("todo: showCoraksNotes") }       // stub: guild.c
func createCharacters()                                          { panic("todo: createCharacters") }      // stub: guild.c
func isQuestItem(id int) int                                     { panic("todo: isQuestItem") }           // stub: items.c
func bashDoor(x, y, dir int)                                     { panic("todo: bashDoor") }              // stub: items.c
func castSpellDialog(arg int) int                                { panic("todo: castSpellDialog") }       // stub: items.c
func clearPartyAttackFlags()                                     { panic("todo: clearPartyAttackFlags") } // stub: spells.c
func charScreen(n int)                                           { panic("todo: charScreen") }            // stub: shop.c
func itemScreen(c *Chr, mode int) *Chr                           { panic("todo: itemScreen") }            // stub: shop.c
func setCharscrHotspots()                                        { panic("todo: setCharscrHotspots") }    // stub: shop.c
func countSkills(c *Chr) int                                     { panic("todo: countSkills") }           // stub: shop.c
func showQuickRef()                                              { panic("todo: showQuickRef") }          // stub: shop.c
