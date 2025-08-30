<div align="center">
    <h1>
        <img src=".github/readme/8Stash Banner.png" alt="8Stash Banner">
        <br/>
        <strong>Your work-in-progress, saved and shareable.</strong>
    </h1>
</div>

### About 8Stash

8Stash is a command-line tool designed to help you quickly save and share your work-in-progress with colleagues or
friends.
Throughout my career as a developer, especially during pair programming sessions, I often needed a fast way to hand over
my current changes when switching drivers. I wanted to automate this process to avoid creating temporary commits or
dealing with patch files, which led to the creation of 8Stash.
This project was also a personal challenge to learn Go by building a practical tool. As the project evolves,
contributions and suggestions to improve its architecture and align it further with Go best practices are highly
encouraged.

The Idea and Goal of this Project is: having a simple command line tool that can quickly stash your current
work-in-progress changes, push them to a remote branch, and share them with others. The tool should be easy to use,
efficient, and integrate seamlessly into a developer's workflow.  
Go was chosen as the language for 8Stash because its simplicity, speed, and portability align perfectly with the
project's goal of creating a reliable and efficient command-line tool for developers.
<h1>
</h1>

### Installation and Usage
You can install the latest release of 8Stash with:
```sh
curl -sL https://raw.githubusercontent.com/TimothySpriegade/8stash/main/scripts/install_latest.sh | bash
```
Alternatively, build and install locally:

1. **Clone the repository:**
   ```sh
   git clone git@github.com:TimothySpriegade/8stash.git
    ```

2. **Install Go dependencies:**
   ```sh
   go mod tidy
   ```
   
2. **Use the local install script**
    ```sh
    cd 8stash/scripts
    chmod +x scripts/build_install_locally.sh
    ./scripts/build_install_locally.sh
    ```

After installation, you can use 8Stash from the command line:
```sh
8stash [command] [options]
```
Use the 'help' command for more detailed usage instructions.
```sh
8stash help
```
<h1>
</h1>

### Contributing & Local Setup

Contributions are welcome. 8Stash aims to be a simple and reliable CLI for sharing work in progress.
Please open an issue to discuss larger changes, keep pull requests small and focused, add or update tests for new
behavior, run go fmt and go test ./... before pushing, and update README.md when user facing changes are introduced.
Check open issues and labels to find a good first task.

1. **Clone the repository:**
   ```sh
   git clone git@github.com:TimothySpriegade/8stash.git
    ```

2. **Install Go dependencies:**
   ```sh
   go mod tidy
   ```

3. **Ensure that the project can be built**
   ```sh
    cd 8stash/cmd/8stash
    go build
    ```
   or
    ```sh
    cd 8stash/scripts
    chmod +x scripts/build_install_locally.sh
    ./scripts/build_install_locally.sh
    ```
   This script can also be used during development to install the latest local version of 8Stash directly.

- Ensure the install target (typically ~/bin or ~/.local/bin) is on your PATH:
    ```sh
    export PATH="$HOME/bin:$HOME/.local/bin:$PATH"
    ```

- Make sure the tests are working locally.
    ```sh
    go test ./...
    ```

Thanks for contributing to 8Stash—every issue, discussion, and PR helps make it better. Happy hacking!
<h1>
</h1>

