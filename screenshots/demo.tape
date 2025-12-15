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

# Clear search and go to detail
Sleep 1s
Enter
Sleep 1s
Screenshot screenshots/detail-view.png

# Calendar view
Escape
Sleep 1s
Type "c"
Sleep 1s
Screenshot screenshots/calendar-list-view.png

# TLD view
Escape
Sleep 1s
Type "t"
Sleep 1s
Enter
Sleep 1s
Screenshot screenshots/tld-view.png

# Search
Escape
Type "/"
Sleep 1s
Type "cloud"
Sleep 1s
Screenshot screenshots/search-results.png

# Quit
Type "q"
Sleep 500ms
