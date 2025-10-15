### Nixopus Extensions - Current Supported Spec

#### Execution Phases
- run: main steps
- validate: post-check steps

#### Variables
- Refer with `{{ variable_name }}` inside properties
- Special: `{{ uploaded_file_path }}` when a file is sent via multipart to the run endpoint

#### Step Types

- command
  - properties:
    - cmd: string (required) — shell command to execute
  - notes: variables interpolated; runs over SSH

- file
  - properties:
    - action: string (required) — one of: `move`, `copy`, `upload`, `delete`, `mkdir`
    - src: string — source path (required for move/copy/upload)
    - dest: string — destination path (required for move/copy/upload; target for delete/mkdir)
  - behavior:
    - move: SFTP rename
    - copy: remote `cp -r`
    - upload: SFTP upload from local API host path to remote dest
    - delete: SFTP remove
    - mkdir: SFTP mkdir -p

- service
  - properties:
    - name: string (required)
    - action: string (required) — e.g. `start`, `stop`, `restart`, `enable`, `disable`
  - behavior: prefers `systemctl`; falls back to `service`

- user
  - properties:
    - action: string (required) — one of: `ensure`, `modify`, `delete`, `check`, `add_groups`, `remove_groups`
    - username: string (required)
    - shell: string (optional) — e.g. `/bin/bash`
    - home: string (optional) — e.g. `/home/deploy`
    - groups: string (optional) — comma-separated groups, e.g. `sudo,docker`
  - behavior:
    - ensure: creates user if missing, then applies shell/home/groups
    - modify: applies provided shell/home/groups
    - delete: removes user and home (`userdel -r`)
    - check: prints `exists` or `missing` to step output
    - add_groups: adds user to each group listed in `groups`
    - remove_groups: removes user from each group listed in `groups`

#### Common Options
- ignore_errors: boolean (optional) — on failure, continue to next step
- timeout: number (optional, seconds) — applied to command/service steps via `timeout`

#### Example

```yaml
execution:
  run:
    - name: "Upload file to remote"
      type: "file"
      properties:
        action: "upload"
        src: "{{ uploaded_file_path }}"
        dest: "/tmp/remote.bin"

    - name: "Restart Nginx"
      type: "service"
      properties:
        name: "nginx"
        action: "restart"

  validate:
    - name: "Check file exists"
      type: "command"
      properties:
        cmd: "test -f /tmp/remote.bin"
```
