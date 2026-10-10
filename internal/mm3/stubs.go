package mm3

// Stubs for the callees of the ported modules that are not ported yet, with their final
// signatures. Whoever ports the real one deletes its stub.

func showMessage(msg []byte)                       { panic("todo: showMessage") }         // stub: charinfo.c
func innRosterScreen()                             { panic("todo: innRosterScreen") }     // stub: charinfo.c
func drawCharsheetCursor(n int)                    { panic("todo: drawCharsheetCursor") } // stub: charinfo.c
func setCombatButtons()                            { panic("todo: setCombatButtons") }    // stub: charinfo.c
func pushButtons()                                 { panic("todo: pushButtons") }         // stub: charinfo.c
func popButtons()                                  { panic("todo: popButtons") }          // stub: charinfo.c
func setAnykeyButtons()                            { panic("todo: setAnykeyButtons") }    // stub: charinfo.c
func optionsMenu(intro int)                        { panic("todo: optionsMenu") }         // stub: mazeload.c
func generateRandomItem(lvl int, c *Chr, slot int) { panic("todo: generateRandomItem") }  // stub: mazeload.c
func gameMainLoop() int                            { panic("todo: gameMainLoop") }        // stub: town.c
func chargeHirelings()                             { panic("todo: chargeHirelings") }     // stub: town.c
func addBankInterest()                             { panic("todo: addBankInterest") }     // stub: town.c
func classCanUseItem(cls, id, quiet int) int       { panic("todo: classCanUseItem") }     // stub: shop.c
func calcMaxHp(c *Chr) int                         { panic("todo: calcMaxHp") }           // stub: shop.c
func getConditionStatMod(c *Chr, stat int) int     { panic("todo: getConditionStatMod") } // stub: events.c
func getBonusStatIndex(v int) int                  { panic("todo: getBonusStatIndex") }   // stub: events.c
func getElementIndex(v int) int                    { panic("todo: getElementIndex") }     // stub: events.c
func moveMonsters()                                { panic("todo: moveMonsters") }        // stub: events.c
func playSfx(v int)                                { panic("todo: playSfx") }             // stub: events.c
func charSavingThrow(c *Chr, kind int) int         { panic("todo: charSavingThrow") }     // stub: combat.c
func monsterSavingThrow(lvl int) int               { panic("todo: monsterSavingThrow") }  // stub: combat.c
func monsterTurn(mt int)                           { panic("todo: monsterTurn") }         // stub: combat.c
func refreshView()                                 { panic("todo: refreshView") }         // stub: scene.c
func getCellWalls(x, y, mask int) int              { panic("todo: getCellWalls") }        // stub: scene.c
func getCellWallsRegion(x, y, mask int) int        { panic("todo: getCellWallsRegion") }  // stub: scene.c
func markCellVisited(x, y int)                     { panic("todo: markCellVisited") }     // stub: scene.c
func updateDarkness()                              { panic("todo: updateDarkness") }      // stub: scene.c
func loadSky(n int)                                { panic("todo: loadSky") }             // stub: cpanel.c
func enterMaze(n int)                              { panic("todo: enterMaze") }           // stub: cpanel.c
func clearTempEffects(keep int)                    { panic("todo: clearTempEffects") }    // stub: cpanel.c
func drawCompassButton(k int)                      { panic("todo: drawCompassButton") }   // stub: items.c
func showPleaseWait()                              { panic("todo: showPleaseWait") }      // stub: items.c
func bounceDice()                                  { panic("todo: bounceDice") }          // stub: guild.c
func randomRange(a, b int) int                     { panic("todo: randomRange") }         // stub: sys.asm
