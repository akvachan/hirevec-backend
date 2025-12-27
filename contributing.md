## Tasks tracking system
In order to create a task:
1. Go to a place in code where something needs to be done and type: `\\ TODO ./todos/DDMMYY-hhmmss-Title.md`.
2. `cp` from `./todos/templates/Todo.md` to `./todos/DDMMYY-hhmmss-Title.md`.
3. `open ./todos/DDMMYY-hhmmss-Title.md`
4. Fill out info in `{{}}`
5. After task is done, set `status: Done` and remove the comment from the code.

## Installation
- Development setup with a hot-reload:
```
make watch
```

## Guidelines
- Avoid `text/template`.
- Never use closers, interfaces or contexts unless there is no other way to do what needs to be done.
- Do not download, install, use 3rd party dependencies beside those that are already available in the [go.mod](./go.mod).
