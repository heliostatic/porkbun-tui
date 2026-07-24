# Porkbun TUI Screenshot Generator
# Run with: vhs screenshots/demo.tape

Set Shell bash
Set Width 800
Set Height 600
Set FontSize 14
Set Padding 10

# Domain list
Type "./porkbun-tui --demo"
Enter
Sleep 3s
Screenshot screenshots/domain-list.png

# Detail view. The Down/Up wiggle forces distinct frames before Enter —
# without it, vhs resolves this Screenshot to the pre-Enter frame.
Sleep 1s
Down
Sleep 500ms
Up
Sleep 500ms
Enter
Sleep 2s
Screenshot screenshots/detail-view.png

# Calendar view
Escape
Sleep 2s
Type "c"
Sleep 2s
Screenshot screenshots/calendar-list-view.png

# TLD view
Escape
Sleep 2s
Type "t"
Sleep 2s
Enter
Sleep 2s
Screenshot screenshots/tld-view.png

# Search
Escape
Type "/"
Sleep 2s
Type "cloud"
Sleep 2s
Screenshot screenshots/search-results.png

# Availability checker (canned demo results)
Escape
Escape
Sleep 500ms
Type "a"
Sleep 2s
Type "coolstartup.dev"
Enter
Sleep 2s
Type "porkbun.com"
Enter
Sleep 2s
Type "myawesomeidea.xyz"
Enter
Sleep 2s
Screenshot screenshots/availability-view.png

# Quit (esc first: q types into the availability input)
Escape
Sleep 500ms
Type "q"
Sleep 500ms
