# ccss

Browse your Claude Code sessions and get the exact command to resume any of them.

<!-- Sessions list. See screenshots/README.md for what to capture. -->
![The sessions list, showing each session's first prompt, size and resume command](screenshots/sessions.png)

---

## Install

Needs [Go](https://go.dev/dl/) 1.21+.

### Option A — one command

```bash
go install github.com/p32929/ccss@latest
```

### Option B — from source

**1. Clone the repo**

```bash
git clone https://github.com/p32929/ccss.git
```

**2. Go into it**

```bash
cd ccss
```

**3. Run the script**

```bash
./run.sh
```

That builds it, installs `ccss` globally, and starts it — all in one step. From then on you can just type `ccss` anywhere.

`run.sh` always starts the app in the folder *you* ran it from, not in the repo, so you still land on your own project's sessions.

### Then — make sure it's on your PATH

**1. Check**

```bash
command -v ccss || echo "not on PATH"
```

If it prints a path, you're done.

**2. If it said `not on PATH`, add Go's bin directory**

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc && source ~/.zshrc
```

Use `~/.bashrc` if you're on bash. `run.sh` prints this same line for you if the folder isn't on your PATH.

**3. Check again**

```bash
command -v ccss
```

### Updating and uninstalling

| | Option A | Option B (source) |
|---|---|---|
| Update | `go install github.com/p32929/ccss@latest` | `git pull && ./run.sh` |
| Uninstall | `rm "$(go env GOPATH)/bin/ccss"` | `rm "$(go env GOPATH)/bin/ccss"` and delete the clone |

---

## Use

**1. Run it in a project**

```bash
cd ~/dev/my-app
ccss
```

**2. Choose where to start**

It always asks, so you're never guessing which list you're looking at:

```
 ccss — where do you want to start?

 t  this folder   /Users/you/dev/my-app
                  5 sessions · 1.4MB

 a  all projects  12 projects · 4.2GB

 t this folder · a all projects · enter this folder · q quit
```

<!-- The start chooser. -->
![The start screen, offering this folder's sessions or all projects](screenshots/start.png)

- **`t`** — just this folder's sessions
- **`a`** — every project you've used Claude Code in
- **`enter`** — takes the obvious one (this folder)
- **`q`** — quit

If the folder has no history, it says so, and `t` isn't offered:

```
 –  this folder   /Users/you/scratch
                  no Claude Code sessions in this folder

 a  all projects  12 projects · 4.2GB

 a all projects · enter all projects · q quit
```

<!-- The start chooser when the folder has no sessions. -->
![The start screen saying this folder has no sessions](screenshots/start-empty.png)

**3. Pick a session**

`↑`/`↓` to move, `/` to filter, `s` to re-sort, `enter` to read the transcript.

<!-- A conversation: prompts, replies, thinking, tool calls. -->
![A session transcript with the current prompt pinned at the top](screenshots/transcript.png)

**4. Copy the command**

Press `c`. You get `cd /Users/you/dev/my-app && claude --resume 0f8e2a91-…` — the `cd` is included because `claude --resume` only finds a session from its own directory.

**5. Paste and run it.**

---

## Keys

| Screen | Keys |
|---|---|
| Projects | `↑/↓` move · `/` filter · `s` sort · `enter` open · `p` open a path · `d` delete project · `q` quit |
| Sessions | `↑/↓` move · `/` filter · `s` sort · `enter` read · `c` copy · `m` mode · `d` delete · `esc` back · `q` quit |
| Transcript | `↑/↓ pgup/pgdn` scroll · `g`/`G` ends · `[`/`]` prev/next prompt · `/` search · `n`/`N` matches · `c` copy · `m` mode · `d` delete · `esc` back · `q` quit |

Every screen shows its keys in the footer. `q` quits from anywhere, `esc` goes back one screen.

---

## Resume modes

`m` cycles the command through Claude Code's permission modes:

| Mode | Command shown |
|---|---|
| normal | `claude --resume <id>` |
| plan | `claude --resume <id> --permission-mode plan` |
| accept edits | `claude --resume <id> --permission-mode acceptEdits` |
| auto | `claude --resume <id> --permission-mode auto` |
| don't ask | `claude --resume <id> --permission-mode dontAsk` |
| bypass permissions | `claude --resume <id> --dangerously-skip-permissions` |

The app never runs anything — it only shows you the command. Your mode and sort choices are remembered between runs.

---

## Disk usage

Both lists show sizes, and `s` sorts by `size` to put the biggest first.

- Projects: `12 sessions · 182.6MB · last used 1d ago`, with the total across all of them in the status row.
- Sessions: each session's size, with that project's total in the status row.

---

## Deleting

`d` deletes the selected session, or on the projects list, every session in that project. Nothing happens until you confirm:

```
delete ALL 24 sessions (1.2GB) in /Users/you/dev/app? this cannot be undone — y to delete, any other key cancels
```

<!-- The red confirmation footer. -->
![The delete confirmation shown in the footer](screenshots/delete.png)

Only `y` goes through; every other key cancels.

**There is no undo.** This removes the `.jsonl` files from `~/.claude/projects`, and `claude --resume` can't bring back a session whose file is gone.

---

## Where things live

Claude Code stores sessions at `~/.claude/projects/<encoded-folder-path>/<session-id>.jsonl`. `ccss` reads those files and never modifies them — apart from `d`.

Your settings go in `~/.config/ccss/config.json` (`~/Library/Application Support/ccss/` on macOS).

---

## License

MIT
