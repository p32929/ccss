# ccss

Browse your Claude Code sessions and get the exact command to resume any of them.

```
   Sessions in /Users/you/dev/app

 │ add loading spinner and all resume modes
 │ 0f8e2a91 · 185 msgs · 4h ago · 444.3KB

   remember the last selected resume mode across runs
   0f8e2a91 · 148 msgs · 3h ago · 355.5KB

 sort: recent · filter: off · 5 sessions · on disk: 1.4MB
 resume:  claude --resume 0f8e2a91-4c3d-4b7a-9e11-aa22bb33cc44
 mode: normal — permissions work normally (asks before acting)
 ↑/↓ move · enter read · c copy · m mode · s sort · / filter · d delete · esc back · q quit
```

---

## Install

**1. Install it** (needs [Go](https://go.dev/dl/) 1.21+)

```bash
go install github.com/p32929/ccss@latest
```

**2. Check it's on your PATH**

```bash
command -v ccss || echo "not on PATH"
```

**3. If it said `not on PATH`, add Go's bin directory**

```bash
echo 'export PATH="$PATH:$(go env GOPATH)/bin"' >> ~/.zshrc && source ~/.zshrc
```

Use `~/.bashrc` if you're on bash.

Update with the same install command. Uninstall with `rm "$(go env GOPATH)/bin/ccss"`.

---

## Use

**1. Run it in a project**

```bash
cd ~/dev/my-app
ccss
```

You land straight on that project's sessions.

**2. If the folder has no sessions, it asks**

```
 No Claude Code sessions here

 folder  /Users/you/scratch

 Show all 12 projects instead?  (4.2GB on disk)

 y show all projects · n quit · esc quit
```

`y` shows everything, `n` quits. From a project, `esc` gets you to the full list too.

**3. Pick a session**

`↑`/`↓` to move, `/` to filter, `s` to re-sort, `enter` to read the transcript.

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

Only `y` goes through; every other key cancels.

**There is no undo.** This removes the `.jsonl` files from `~/.claude/projects`, and `claude --resume` can't bring back a session whose file is gone.

---

## Where things live

Claude Code stores sessions at `~/.claude/projects/<encoded-folder-path>/<session-id>.jsonl`. `ccss` reads those files and never modifies them — apart from `d`.

Your settings go in `~/.config/ccss/config.json` (`~/Library/Application Support/ccss/` on macOS).

---

## License

MIT
