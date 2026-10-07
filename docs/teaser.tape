Output teaser.gif
Output teaser.mp4
Output teaser.webm

Require task
#Set WindowBar Rings
Set FontSize 12
#Set FontFamily Consolas
Set Columns 60
Set Rows 20
Set Padding 10
Set TypingSpeed 50ms

# intro teaser...
Sleep 500ms
Type "Tried using commitizen/git-cz?"
Sleep 1s
Type " Or the alternatives?"
Sleep 2s
#Backspace@10ms 51
#Sleep 1s
Ctrl+A
Ctrl+K
Type "Find the line-by-line editing kind of"
Type@750ms "..."
Type " primitive?"
Sleep 2s
#Backspace@10ms 51
Ctrl+A
Ctrl+K
#Sleep 1s
Type "Get committed!"
Sleep 2s
Hide
#Backspace@10ms 14
Ctrl+A
Ctrl+K

# build and fake git function
Hide
Set TypingSpeed 10ms
Type "task build"
Enter
Type "echo $SHELL"
Enter
Type "alias"
Enter
Type "git() {"
Enter
#Type "  ../.build/committed -n demo.msg"
Type "  ../.build/committed -n"
Enter
Type "}"
Enter
Type "clear"
Enter
Sleep 10ms

# actual demo
Show
Set TypingSpeed 75ms
Type "git commit"
Enter
Sleep 2
Down@300ms 6
Sleep 0.5
Up@200ms 2
Tab 2
Type "testing committed"
Tab
Type "You just type your message here."
Sleep 1
Type " It knows your commitlint settings (or uses defaults) and can line-wrap for you.
Sleep 1
Type " And if you make a misteak?"
Enter
Type "... oops!"
Sleep 1
Type " It's a *full* editor, so you can go back and fix it!"
Sleep 1
Up
Backspace 3
Type "ake"
Sleep 2
#Ctrl+Enter
Ctrl+c
#Wait
Sleep 2
Type "It's that easy! (Ctrl+Enter in vhs isn't working for me 🤷)"
