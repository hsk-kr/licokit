---
name: pr-comments
description: Check and resolve PR review comments from GitHub
---

Check for pending PR review comments and resolve them.

Steps:

1. Check if current branch has a PR:
   ```bash
   gh pr view --json number,url,title 2>&1
   ```
   If this fails (exit code non-zero), tell the user: "No PR found for the current branch. Push your branch and create a PR first." Then stop.

2. Extract repo owner, name, and PR number:
   ```bash
   OWNER=$(gh repo view --json owner -q '.owner.login')
   REPO=$(gh repo view --json name -q '.name')
   PR_NUMBER=$(gh pr view --json number -q '.number')
   ```

3. Get all review comments:
   ```bash
   gh api "repos/$(gh repo view --json nameWithOwner -q .nameWithOwner)/pulls/${PR_NUMBER}/comments" --jq '.[] | {id: .id, in_reply_to_id: .in_reply_to_id, path: .path, line: .line, body: .body, user: .user.login, created: .created_at}'
   ```

4. Get review threads to check resolution status:
   ```bash
   gh api graphql -F owner="${OWNER}" -F repo="${REPO}" -F number="${PR_NUMBER}" -f query='
     query($owner: String!, $repo: String!, $number: Int!) {
       repository(owner: $owner, name: $repo) {
         pullRequest(number: $number) {
           reviewThreads(first: 100) {
             nodes {
               id
               isResolved
               comments(first: 10) {
                 nodes {
                   id
                   databaseId
                   body
                   author { login }
                   path
                   line
                 }
               }
             }
           }
         }
       }
     }
   '
   ```

5. For each unresolved comment:
   - Read the referenced file and line
   - Understand the reviewer's feedback
   - Make the requested change (or explain why not)
   - After addressing, reply to the comment acknowledging what was done:
     ```bash
     gh api "repos/$(gh repo view --json nameWithOwner -q .nameWithOwner)/pulls/${PR_NUMBER}/comments" \
       -X POST \
       -f body="Done — <brief description of the change made>" \
       -F in_reply_to=<original_comment_id>
     ```
   - Report what was done

6. Summarize all changes made and any comments that need discussion.
