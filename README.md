# SCP-EZ 🚀
🚀 Eazy to use SCP client for transferring files between two systems.

## 🖼 Screenshot️
![screen](https://github.com/tkkwa01/scpez/assets/130450932/42ae415d-255b-43e6-9eed-4562a2d81b72)

## 📝️ Description
This application is a simple SCP client that allows users to transfer files between two systems. The application is built using Go and tview library. The application is cross-platform and can be run on macOS, Linux and FreeBSD.

## 💻 Support OS
- macOS
- Linux
- FreeBSD

## ⬇️ Installation
Clone this repository
   ```sh
   git clone git@github.com:tkkwa01/scpez.git
   ```
   
## 🏃Usage
Run the following command to start the application
```sh
cd scpez

./scpez-mac
or 
./scpez-linux
or 
./scpez-freebsd
```

### Add the application to the PATH
add the following line to the .bashrc or .zshrc file
```sh
export PATH="/path/to/project/scpez:$PATH"
```
and you can run the application from anywhere in the terminal
```sh
scpez-mac
```

## 🧑‍🎓 How to use
First, enter the server name, username, and password. Connect to the server and select the directory or file you want using the space key. Press the T key to copy the selected directory or file to your home directory.
If you transfer directories, the application will create a new directory named `SCP-EZ` in the home directory. The transferred files will be stored in the `SCP-EZ` directory.

##  👩‍💻 Keybindings
| Key        | Description                    |
|------------|--------------------------------|
| `Tab`       | next panes                     |
| `Shft + Tab` | back panes                     |
| `Enter`      | navigate to the directory      |
| `B`          | back to the previous directory |
| `L`          | preview the file               |
| `Q`          | quit preview                   |
| `Space`      | select / unselect file         |
| `T`          | transfer selected files        |
| `ctrl + c`   | quit the application           |
