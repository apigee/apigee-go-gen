# Using Git Templates

The `--template` argument, used in various `apigee-go-gen render` commands, accepts two types of sources for your template files: a local file path or a remote file path via a Git repository URI.

## 1. Local File Path

This is the simplest usage, pointing directly to a file on your local machine.

```
# Relative path example
apigee-go-gen render ... --template ./templates/my_api.yaml.tpl

# Absolute path example
apigee-go-gen render ... --template /home/user/project/templates/my_api.yaml.tpl
```

## 2. Remote Git Repository URI

!!! Info
    To use this remote fetching feature, you need **Git 2.25 or later** installed and available in your `$PATH`.

You can specify a template file located in a public or private Git repository using a structured URI. The tool fetches only the necessary file and its containing directory using a sparse checkout, minimizing download time.

A Git URI is composed of three parts: **Repository URL**, **Reference (Ref)**, and **Resource Path**.

### Git URI Structure

The Git URI must conform to one of two primary formats for parsing: **Web-View** or **Generic**.

| Format       | Purpose                                                             | 
|--------------|---------------------------------------------------------------------| 
| **Web-View** | Standard format seen in GitHub, GitLab, and Bitbucket.              | 
| **Generic**  | A flexible format for any Git URL, including SSH and generic HTTPS. | 

### A. Web-View Format

These formats match the path structure you see when browsing files on major Git hosting platforms.

| Style         | Structure                            | 
|---------------|--------------------------------------| 
| **GitHub**    | `<REPO_URL>/blob/<REF>/<RESOURCE>`   | 
| **GitLab**    | `<REPO_URL>/-/blob/<REF>/<RESOURCE>` | 
| **BitBucket** | `<REPO_URL>/src/<REF>/<RESOURCE>`    | 

#### Example Usage

```
# Using GitHub web-view format with the 'main' branch
apigee-go-gen render ... --template https://github.com/my-org/templates/blob/main/templates/my/api.yaml

# Using GitLab web-view format with a specific tag
apigee-go-gen render ... --template https://gitlab.com/apigee/starter/-/blob/v1.2.3/templates/my/api.yaml

# Using BitBucket web-view format with a commit hash
apigee-go-gen render ... --template https://bitbucket.org/team/project/src/a1b2c3d4e5f6/templates/my/api.yaml
```

### B. Generic Format

This flexible format uses the `/-/` separator and is useful for non-HTTP Git protocols (like SSH) or when you need explicit control over the **Repository URI**, **Reference** and **Resource** .

| Style       | Structure                   | Purpose                                                       | 
|-------------|-----------------------------|---------------------------------------------------------------| 
| **Simple**  | `<REPO>/-/<REF>/<RESOURCE>` | Used when `<REF>` has no slashes (e.g., `main`).              | 
| **Complex** | `<REPO>/-/<REF>#<RESOURCE>` | Used when `<REF>` contains slashes (e.g., `feature/new-api`). | 

#### Example Usage

```
# Using SSH protocol (implicit) with a simple branch name ('main')
apigee-go-gen render ... --template git@github.com:my-org/project.git/-/main/my/api.yaml

# Using SSH protocol (explicit) with a simple branch name ('main')
apigee-go-gen render ... --template ssh://git@github.com:my-org/project.git/-/main/my/api.yaml

# Using HTTPS protocol with a tag ('v1.2.3')
apigee-go-gen render ... --template https://gitlab.com/my-org/project.git/-/v1.2.3/my/api.yaml

# Using a reference with a slash (branch 'feature/jira-1337') requiring '#'
apigee-go-gen render ... --template https://repo.example.com/project.git/-/feature/jira-1337#my/api.yaml
```

### Reference Types

The reference (`<REF>`) component of the URI is flexible and can point to any valid Git object:

* **Branch Name:** The name of a branch (e.g., `main`, `develop`, `feature/new-service`).

* **Tag:** A version tag (e.g., `v1.0.0`, `release-2024-05`).

* **Commit Hash:** A full or short commit hash (e.g., `a1b2c3d4e5f6` or `a1b2c3d`).

## Private Repositories (SSH)

To access private Git repositories, the preferred and most secure mechanism is using **SSH keys**. This requires your SSH client to have the correct key loaded into the `ssh-agent`.

### SSH Key Setup

Before attempting to fetch a private template, ensure you have an SSH key pair generated and added to your Git hosting platform account (GitHub, GitLab, etc.).

**A. Generate an SSH Key Pair**

If you do not have an SSH key, use `ssh-keygen`. We recommend using the default file name (`id_rsa` or `id_ed25519`) and adding a secure passphrase.

```sh
ssh-keygen -t ed25519 -C "your_email@example.com"
```

**B. Add Key to SSH-Agent**

You must start the `ssh-agent` and add your private key so that Git can use it without prompting for your passphrase on every fetch operation.

```sh
# 1. Start the ssh-agent in the background
eval "$(ssh-agent -s)"

# 2. Add your private key (use the path to your generated key)
ssh-add ~/.ssh/id_ed25519
```

### Troubleshooting SSH Connectivity

If you encounter issues when attempting to fetch a template, it is often related to the client's knowledge of the Git host's server key.

**Error: `ssh: handshake failed: knownhosts: key is unknown`**

This indicates that your `~/.ssh/known_hosts` file does not contain the public SSH key for the Git server (e.g., GitHub.com or Bitbucket.org). You need to add it:

```sh
# Replace bitbucket.org with your Git host (e.g., github.com, gitlab.com)
ssh-keyscan -H bitbucket.org >> ~/.ssh/known_hosts
```

**Error: `ssh: handshake failed: knownhosts: key mismatch`**

This means the existing key for the Git SSH server in your `~/.ssh/known_hosts` file is incorrect (e.g., the host changed their key). You must first remove the old key, and then add the new one.

```sh
# 1. Remove the existing, incorrect key
ssh-keygen -R bitbucket.org

# 2. Add the correct host keys
ssh-keyscan -H bitbucket.org >> ~/.ssh/known_hosts
```


## Private Repositories (HTTPS)

If you cannot use SSH, you can access private repositories using the HTTPS URI format by embedding a Personal Access Token (PAT) as the password component.

1.  **Generate a PAT:** Create a Personal Access Token in your Git hosting platform's security settings. This token must have the scope necessary to read repository contents (`repo` or similar).

2.  **Use the Token in the URI:** The token replaces the password in the standard HTTPS URL structure.

    **Structure:** `https://<USERNAME>:<TOKEN>@<GIT_HOST>/<REPO_PATH>`

!!! Warning
    This method is less secure than using SSH, and should be avoided if possible. 
    Your Personal Access Token is briefly written to a temporary `.git` directory during the `git checkout` operation.
    Once the `git checkout` operation completes, the `.git` directory is deleted.

#### Example Usage (GitHub)

```
# Replace 'USERNAME' and 'TOKEN'
apigee-go-gen render --template https://USERNAME:TOKEN@github.com/my-org/templates/-/main/path/to/template.tpl
```