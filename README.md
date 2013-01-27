Screencastinator
================

Screencastinator is a Tool to edit your ANSI console screencasts created with the standard linux-util tool "script".

[![screenshot](http://gbin.github.com/screencastinator/screencastinator.png)](http://gbin.github.com/screencastinator/screencastinator.png)


Making a screencast is extremely tedious. This tool gives you the opportunity to correct the mistakes you made during the recording without having to correct them in a video editor.

For example you made this typo during the recording: % cd mydirrec^H^H^Herctory

Scriptcastinator will help you replay your screencast to the point: % cd mydir and you will have the opportunity to make your correction "rec^H^H^H" disappear.

**Note: this is an early release and expect bugs while using it. Any contributor is welcomed.**

## Downloads ##

_Prerelease v0.1_

[MacOS amd64 binary](http://gbin.github.com/screencastinator/releases/screencastinator-v0.1-darwin-amd64)

[Linux 386 binary](http://gbin.github.com/screencastinator/releases/screencastinator-v0.1-linux-386)

[Linux amd64 binary](http://gbin.github.com/screencastinator/releases/screencastinator-v0.1-linux-amd64)

## Requirements ##
* Go 1.0.3+  (only if you want to contribute)
* linux-util (if you want to record you screencast, not needed for edition)

## How to use it ##

You need to record you screencast with "script", and you need to record a timing file.
Here is a small bash script I use to simplify it (it erases the screen before the start for example)

record.sh:

```BASH
    echo -ne "\e[8;43;132t"
    clear
    script --timing=$1.timing -q $1.session
```

This will create 2 files ending with .timing and .session.

You can pass them on to screencastinator :

```
screencastinator file
```
(screencastinator will add .timing and .session automatically) 

The editor is composed by :
- a view window at the top
- a timeline showing your current position in the stream
- a status area with more explanations about the current ANSI character you are on
- a quick keyboard help

## Tips ##

The [n] key allows you to automatically extend the current selection to the time where your cursor is back at the same place. Basically it autodetect the blahblah^H^H^H pattern for you. 

