# Porkbun TUI Screenshot Generator
# Run with: vhs screenshots/demo.tape

Set Shell bash
Set Width 100
Set Height 28
Set FontSize 14
Set Padding 10

# Domain list
Type "./porkbun-tui --demo"
Enter
Sleep 1.5s
Screenshot screenshots/domain-list.png

# Search
Type "/"
Sleep 300ms
Type "cloud"
Sleep 500ms
Screenshot screenshots/search-results.png

# Clear search and go to detail
Escape
Sleep 300ms
Enter
Sleep 500ms
Screenshot screenshots/detail-view.png

# TLD view
Escape
Sleep 300ms
Type "t"
Sleep 500ms
Enter
Sleep 300ms
Screenshot screenshots/tld-view.png

# Calendar view
Escape
Sleep 300ms
Type "c"
Sleep 500ms
Screenshot screenshots/calendar-list-view.png

# Quit
Type "q"
Sleep 300ms
