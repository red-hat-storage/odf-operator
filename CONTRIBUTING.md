# How to Contribute

## Commit structure

ODF maintainers value clear and explanatory commit messages. So by default each of your commits must follow below rules:

### We follow the common commit conventions
```
type: subject

body?

footer?
```

### Here is an example of an acceptable commit message for a bug fix:
```
component: commit title

This is the commit message, here I'm explaining, what the bug was along
with its root cause.
Then I'm explaining how I fixed it.

Fix: https://bugzilla.redhat.com/show_bug.cgi?id=<NUMBER>

Signed-off-by: First_Name Last_Name <email address>
```

### Here is an example of an acceptable commit message for a new feature:
```
component: commit title

This is the commit message, here I'm explaining, what this feature is
and why do we need it.

Signed-off-by: First_Name Last_Name <email address>
```

### type/component must be one of the following according to files you are changing:
```yaml
action
api
bundle
ci
controllers
docs
godeps
makefile
test
webhook
```

Note: sometimes you will feel like there is not so much to say, for instance if you are fixing a typo in a text. In that case, it is acceptable to shorten the commit message.

### More Guidelines:
- Type/component should not be empty.
- Your commit msg should not exceed more than 72 characters per line.
- Header should not have a full stop.
- Body should always end with the full stop.
- There should be one blank line in b/w header and body.
- There should be one blank line in b/w body and footer.
- Your commit message must be signed-off.
