---
outline: deep
---

# Terminal

find out all the configurations and features that Terminal offers.

## Key Bindings
The following key bindings are available in the terminal:

* `CTRL + J`: Toggle the terminal on and off.
* `CTRL + T`: Change the position of the terminal to the bottom or right side of the application.

## Resize
You can resize the terminal by dragging the terminal from the top, which provides the same behaviour as vscode's terminal.

## Themes and fonts
The terminal inherits the theme and font from the application. See [Theme and Fonts](/themes-and-fonts) for more information on how to change them.

## Terminal focus
When you click inside the terminal, some keybindings won't work for the application anymore because they are treated as terminal's keybindings. 
For example, `CTRL + C` won't copy the file, instead it sends a SIGINT to the running command.

## Coming Soon :rocket:
* Prompt customization
* Support for all the editors
* Detection of system nuking commands and privacy protections
* Switching the terminal between multiple servers (master/slaves)
* Auto completion
* AI powered suggestions / corrections of the commands
* Support for multiple shells switching (bash/zsh/fish)
