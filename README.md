[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![GitHub Release](https://img.shields.io/github/v/release/timothyspriegade/8stash)
![GitHub repo size](https://img.shields.io/github/repo-size/timothyspriegade/8stash)

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

### Installation
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

<h1>
</h1>

### Usage
8Stash lets you quickly park your uncommitted work into a temporary remote branch, pick it up on another machine, and re‑apply it as local unstaged changes without polluting your main branch history.

Prerequisites:
1. You are inside a Git repository with a clean syncable base (no unpushed divergent commits)
2. You have uncommitted changes you want to share or move
3. origin remote is available (SSH or HTTPS, preferably SSH)

Typical workflow:
1. Save (push): Create a temporary branch from the current HEAD that contains exactly your current uncommitted changes; they are committed on that new branch, the branch is pushed to origin, and your original branch is restored to a clean state locally.
   - Optionally add a descriptive message to your stash using the `-m` flag: `8stash push -m "implementing login feature"`
   - The commit message will be displayed when listing stashes, making it easier to identify specific work-in-progress items
2. Share: Communicate the temporary branch number (e.g. 8stash/8374) to a teammate or switch machines and pull/fetch on the other clone.
3. List (list): View available stash branches filtered by a naming convention (e.g. all starting with 8stash/) with human‑readable ages, author information, and commit messages to choose the right one.
4. Apply (pop): Re‑apply a chosen stash branch onto your current branch so that its changes appear as unstaged modifications in your working directory; no merge commit and no history rewrite occur.
5. Clean up: Applied stashes are removed localy and on remote to avoid messy repositories.
6. Work: Inspect, edit further, stage, and create proper commits as desired.
7. Repeat as needed for new slices of in‑progress work.

Behavior characteristics:
- Divergent histories are now supported: you can apply stashes even if your current branch has diverged from the stash base. The tool will attempt to merge changes, and you may need to resolve conflicts manually.
- Applying a stash does not advance or modify your current branch's commit history; it only repopulates the working tree.
- Relative age displays (e.g. minutes/hours/days ago) are based on the current system clock.
- Stash commit messages help you identify and organize your work-in-progress items across different contexts.

Ideal use cases:
- Pair programming driver handoff.
- Moving unfinished edits between desktop and laptop.
- Quick review handover without committing partial or experimental changes.
- Temporary parking of exploratory work prior to reshaping into clean commits.
- Annotating work-in-progress with descriptive messages for better organization.

Limitations to keep in mind:
- Not a replacement for long‑lived feature branches.
- Does not resolve underlying repository divergence; you must reconcile first.
- Large binary or generated assets will still be committed onto the temporary branch (consider .gitignore hygiene).

Use the 'help' command for further detailed usage instructions.
```sh
8stash help
```

#### Command Examples

**Push with a descriptive message:**
```sh
8stash push -m "WIP: refactoring user authentication"
```

**Push without a message (uses default):**
```sh
8stash push
# or simply
8stash
```

**List all stashes (shows messages, authors, and timestamps):**
```sh
8stash list
```

**Pop a specific stash:**
```sh
8stash pop 8374
```

**Drop a stash you no longer need:**
```sh
8stash drop 8374
```

**Clean up old stashes:**
```sh
8stash cleanup
# or skip confirmation prompt
8stash cleanup -y
# or override retention period
8stash cleanup -d 7
```

<h1>
</h1>

### Optional Configuration

8Stash can be configured via a `.8stash.yaml` file placed in your repository's root directory. This allows you to customize behavior like branch naming and cleanup policies on a per-project basis.

If no `.8stash.yaml` is found, the application will use its default settings.

**Example `.8stash.yaml`:**

```yaml
branch_prefix: "project-wip/"
retention_days: 15
naming:
  hash_type: "uuid" # "numeric" or "uuid"
  hash_numeric_max_value: 99999
```

#### Configuration Options

| Key                        | Type   | Description                                                                                             | Default      |
| -------------------------- | ------ | ------------------------------------------------------------------------------------------------------- | ------------ |
| `branch_prefix`            | string | The prefix for all stash branches created by 8Stash. A trailing `/` is added automatically.             | `8stash/`    |
| `retention_days`           | int    | The number of days after which a stash is considered "old" and eligible for the `cleanup` command.      | `30`         |
| `naming.hash_type`         | string | The format for generated stash IDs. Must be either `"numeric"` or `"uuid"`.                             | `"numeric"`  |
| `naming.hash_numeric_max_value` | int    | The exclusive upper bound for randomly generated numeric stash IDs (e.g., a value of `10000` generates IDs from 0-9999). | `9999`       |

**Notes:**
*   The `retention_days` value can be temporarily overridden for a single run by using the `-d` or `--days` flag on the `cleanup` command (e.g., `8stash cleanup -d 10`).
*   Confirmation prompts for the `cleanup` command can be skipped by using the `-y` or `--yes` flag (e.g., `8stash cleanup -y`).
*   If `naming.hash_type` is set to `"uuid"`, the `hash_numeric_max_value` is ignored.
*   The application will print a warning and clamp the value if `hash_numeric_max_value` is set above the maximum supported value (`2,147,483,647`).
*   Invalid values in the config file will cause the application to fall back to the default settings for that specific key.

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
<h1>
</h1>

### Contributers
<a href="https://github.com/timothyspriegade/8stash/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=timothyspriegade/8stash" />
</a>

Thanks for contributing to 8Stash—every issue, discussion, and PR helps make it better. Happy hacking!
<h1>
</h1>

### Disclaimer
This project is my first experience with Go, so the code quality and structure may not be optimal. I am continuously learning and improving, and any feedback, suggestions, or contributions are always welcome and appreciated!
<h1>
</h1>


