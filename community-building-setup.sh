#!/usr/bin/env bash
# Community Building Script for ACM Project
# This script sets up all community infrastructure and templates

set -euo pipefail

# Configuration
REPO_OWNER="acm-project"
REPO_NAME="acm"
PROJECT_DIR="$(pwd)"

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  ACM Community Building Script                          ║${NC}"
echo -e "${BLUE}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""

# Check prerequisites
echo -e "${BLUE}[1/10] Checking prerequisites...${NC}"
if ! command -v gh &> /dev/null; then
    echo -e "${YELLOW}GitHub CLI required. Install: brew install gh${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Prerequisites met${NC}"

# Create directory structure
echo -e "${BLUE}[2/10] Creating directory structure...${NC}"
mkdir -p .github/{ISSUE_TEMPLATE,workflows,DISCUSSION_TEMPLATE}
mkdir -p docs/{guides,architecture,legal,community}
mkdir -p scripts/community
mkdir -p .github/pull_request_template
echo -e "${GREEN}✓ Directory structure created${NC}"

# Create CONTRIBUTING.md
echo -e "${BLUE}[3/10] Creating CONTRIBUTING.md...${NC}"
cat > CONTRIBUTING.md << 'EOF'
# Contributing to ACM

Thank you for your interest in contributing to Automated Compromise Mitigation (ACM)! 🎉

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Workflow](#development-workflow)
- [Code Quality Standards](#code-quality-standards)
- [Commit Message Guidelines](#commit-message-guidelines)
- [Pull Request Process](#pull-request-process)
- [Community](#community)

## Code of Conduct

ACM adopts the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold this code.

## Getting Started

### Prerequisites

- **Go 1.21+** (for core service development)
- **Node.js 18+** (for Tauri GUI)
- **Rust 1.70+** (for Tauri backend)
- **Password Manager CLI**: 1Password or Bitwarden installed and configured

### Setup Development Environment

```bash
# Clone repository
git clone https://github.com/acm-project/acm.git
cd acm

# Install dependencies
make deps

# Run tests
make test

# Run linters
make lint

# Build
make build
```

## How to Contribute

We welcome contributions in many forms:

### 🐛 Bug Reports

Found a bug? [Open an issue](https://github.com/acm-project/acm/issues/new?template=bug_report.md) with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, etc.)

### ✨ Feature Requests

Have an idea? [Open an issue](https://github.com/acm-project/acm/issues/new?template=feature_request.md) with:
- Use case and problem statement
- Proposed solution
- Alternatives considered

### 📝 Documentation

- Fix typos or unclear sections
- Add examples or tutorials
- Translate documentation to other languages
- Improve API documentation

### 💻 Code Contributions

See [good first issues](https://github.com/acm-project/acm/labels/good%20first%20issue) for beginner-friendly tasks.

## Development Workflow

### 1. Find or Create Issue

- Browse [open issues](https://github.com/acm-project/acm/issues)
- Comment on issue to indicate you're working on it
- If no issue exists, create one first (avoids duplicate work)

### 2. Fork and Branch

```bash
# Fork repository on GitHub, then:
git remote add upstream https://github.com/acm-project/acm.git
git checkout -b feature/your-feature-name
```

**Branch Naming Convention:**
- `feature/description` — new features
- `bugfix/description` — bug fixes
- `docs/description` — documentation updates
- `chore/description` — maintenance tasks

### 3. Develop

```bash
# Make changes
# Add tests
make test

# Run linters
make lint

# Ensure build succeeds
make build
```

### 4. Commit

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```bash
git commit -m "feat(crs): add LastPass CLI integration"
git commit -m "fix(audit): prevent race condition in log writer"
git commit -m "docs: update installation guide with Windows instructions"
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Create PR on GitHub with:
- Clear description of changes
- Link to related issue (`Closes #123`)
- Screenshots (if UI changes)
- Testing notes

## Code Quality Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go)
- Run `go fmt` before committing
- Use `golangci-lint` with project configuration
- Maximum cyclomatic complexity: 15
- Comment all exported functions and types (GoDoc)

**Example:**

```go
// RotateCredential rotates a compromised credential by generating a new
// secure password and updating the vault entry via the password manager CLI.
//
// Returns an error if:
//   - Password generation fails
//   - Vault update fails
//   - Verification of updated credential fails
func (crs *CredentialRemediationService) RotateCredential(
    ctx context.Context,
    cred CompromisedCredential,
) error {
    // Implementation...
}
```

### TypeScript/React Code Style

- Use Prettier for formatting
- ESLint with project configuration
- TypeScript strict mode enabled
- Functional components with hooks (no class components)

### Testing Requirements

- **Unit tests:** > 80% coverage for new code
- **Integration tests:** For CLI interactions, API endpoints
- **E2E tests:** For critical user workflows (Tauri GUI)

```bash
# Run tests with coverage
make test-coverage

# View coverage report
open coverage.html
```

## Commit Message Guidelines

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- **feat:** New feature
- **fix:** Bug fix
- **docs:** Documentation changes
- **style:** Code style changes (formatting, no logic change)
- **refactor:** Code refactoring
- **test:** Add or update tests
- **chore:** Maintenance tasks (dependencies, build scripts)
- **perf:** Performance improvements
- **ci:** CI/CD changes

### Scopes

- `crs` — Credential Remediation Service
- `acvs` — Automated Compliance Validation Service
- `him` — Human-in-the-Middle Manager
- `tui` — OpenTUI client
- `gui` — Tauri GUI
- `audit` — Audit logging
- `security` — Security-related changes
- `legal` — Legal/compliance changes
- `docs` — Documentation

### Examples

```
feat(crs): add support for KeePassXC CLI integration

Implements KeePassXC adapter following the abstraction pattern
used for 1Password and Bitwarden.

Closes #87

---

fix(audit): prevent race condition in concurrent log writes

Added mutex to protect SQLite writes during concurrent rotation
operations. Includes regression test.

Fixes #142

---

docs(contributing): add section on commit message guidelines

Clarifies expected commit format and provides examples.
```

## Pull Request Process

### Before Submitting PR

- [ ] All tests pass (`make test`)
- [ ] Linters pass (`make lint`)
- [ ] Coverage meets requirements (> 80% for new code)
- [ ] Documentation updated (if API changed)
- [ ] CHANGELOG entry added (if user-facing change)

### PR Template

When creating a PR, fill out the template:

```markdown
## Description
[Brief description of changes]

## Related Issue
Closes #[issue number]

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Testing
[Describe testing performed]

## Screenshots (if applicable)
[Add screenshots for UI changes]

## Checklist
- [ ] My code follows the style guidelines
- [ ] I have performed a self-review
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have updated the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally
```

### Review Process

1. **Automated Checks:** CI/CD must pass
2. **Code Review:** Requires 2+ Core Maintainer approvals
3. **Security Review:** Security Lead reviews security-critical changes
4. **Legal Review:** Legal Review Committee reviews EULA/compliance changes

**Review Timeline:**
- Trivial changes: 24 hours
- Minor changes: 48 hours
- Major changes: 1 week
- Critical/breaking changes: 2 weeks (with RFC if needed)

### After Approval

- Maintainer will squash and merge
- Your contribution will be credited in release notes
- Issue automatically closed via `Closes #123` in commit message

## Community

### Communication Channels

- **GitHub Discussions:** https://github.com/acm-project/acm/discussions
- **Discord:** https://discord.gg/acm (real-time chat)
- **Monthly Community Call:** First Thursday of each month, 5pm UTC
- **Twitter/X:** @acm_project

### Getting Help

- **Questions:** Use GitHub Discussions or #help channel on Discord
- **Bugs:** Open a GitHub Issue
- **Security Issues:** Email security@acm.dev (do NOT open public issue)

### Recognition

We celebrate contributors!

- **First PR merged:** GitHub badge, Discord role, featured in newsletter
- **Contributor of the Month:** Blog post spotlight, special Discord role
- **Core Contributor (10+ PRs):** Listed on website, invited to Core Contributor sync calls

---

## Thank You!

Your contributions make ACM better for everyone. We appreciate your time and effort! 🙏

If you have questions, don't hesitate to ask in [GitHub Discussions](https://github.com/acm-project/acm/discussions) or [Discord](https://discord.gg/acm).

**Let's build something great together!** 🚀
EOF
echo -e "${GREEN}✓ CONTRIBUTING.md created${NC}"

# Create CODE_OF_CONDUCT.md
echo -e "${BLUE}[4/10] Creating CODE_OF_CONDUCT.md...${NC}"
cat > CODE_OF_CONDUCT.md << 'EOF'
# Contributor Covenant Code of Conduct

## Our Pledge

We as members, contributors, and leaders pledge to make participation in our
community a harassment-free experience for everyone, regardless of age, body
size, visible or invisible disability, ethnicity, sex characteristics, gender
identity and expression, level of experience, education, socio-economic status,
nationality, personal appearance, race, caste, color, religion, or sexual
identity and orientation.

We pledge to act and interact in ways that contribute to an open, welcoming,
diverse, inclusive, and healthy community.

## Our Standards

Examples of behavior that contributes to a positive environment for our
community include:

* Demonstrating empathy and kindness toward other people
* Being respectful of differing opinions, viewpoints, and experiences
* Giving and gracefully accepting constructive feedback
* Accepting responsibility and apologizing to those affected by our mistakes,
  and learning from the experience
* Focusing on what is best not just for us as individuals, but for the overall
  community

Examples of unacceptable behavior include:

* The use of sexualized language or imagery, and sexual attention or advances of
  any kind
* Trolling, insulting or derogatory comments, and personal or political attacks
* Public or private harassment
* Publishing others' private information, such as a physical or email address,
  without their explicit permission
* Other conduct which could reasonably be considered inappropriate in a
  professional setting

## Enforcement Responsibilities

Community leaders are responsible for clarifying and enforcing our standards of
acceptable behavior and will take appropriate and fair corrective action in
response to any behavior that they deem inappropriate, threatening, offensive,
or harmful.

Community leaders have the right and responsibility to remove, edit, or reject
comments, commits, code, wiki edits, issues, and other contributions that are
not aligned to this Code of Conduct, and will communicate reasons for moderation
decisions when appropriate.

## Scope

This Code of Conduct applies within all community spaces, and also applies when
an individual is officially representing the community in public spaces.

## Enforcement

Instances of abusive, harassing, or otherwise unacceptable behavior may be
reported to the community leaders responsible for enforcement at
community@acm.dev.

All complaints will be reviewed and investigated promptly and fairly.

All community leaders are obligated to respect the privacy and security of the
reporter of any incident.

## Enforcement Guidelines

Community leaders will follow these Community Impact Guidelines in determining
the consequences for any action they deem in violation of this Code of Conduct:

### 1. Correction

**Community Impact**: Use of inappropriate language or other behavior deemed
unprofessional or unwelcome in the community.

**Consequence**: A private, written warning from community leaders, providing
clarity around the nature of the violation and an explanation of why the
behavior was inappropriate. A public apology may be requested.

### 2. Warning

**Community Impact**: A violation through a single incident or series of
actions.

**Consequence**: A warning with consequences for continued behavior. No
interaction with the people involved, including unsolicited interaction with
those enforcing the Code of Conduct, for a specified period of time. This
includes avoiding interactions in community spaces as well as external channels
like social media. Violating these terms may lead to a temporary or permanent
ban.

### 3. Temporary Ban

**Community Impact**: A serious violation of community standards, including
sustained inappropriate behavior.

**Consequence**: A temporary ban from any sort of interaction or public
communication with the community for a specified period of time. No public or
private interaction with the people involved, including unsolicited interaction
with those enforcing the Code of Conduct, is allowed during this period.
Violating these terms may lead to a permanent ban.

### 4. Permanent Ban

**Community Impact**: Demonstrating a pattern of violation of community
standards, including sustained inappropriate behavior, harassment of an
individual, or aggression toward or disparagement of classes of individuals.

**Consequence**: A permanent ban from any sort of public interaction within the
community.

## Attribution

This Code of Conduct is adapted from the [Contributor Covenant][homepage],
version 2.1, available at
[https://www.contributor-covenant.org/version/2/1/code_of_conduct.html][v2.1].

[homepage]: https://www.contributor-covenant.org
[v2.1]: https://www.contributor-covenant.org/version/2/1/code_of_conduct.html
EOF
echo -e "${GREEN}✓ CODE_OF_CONDUCT.md created${NC}"

# Create GitHub Issue Templates
echo -e "${BLUE}[5/10] Creating GitHub Issue templates...${NC}"

# Bug Report Template
cat > .github/ISSUE_TEMPLATE/bug_report.md << 'EOF'
---
name: Bug Report
about: Report a bug to help us improve ACM
title: '[BUG] '
labels: 'bug, needs-triage'
assignees: ''
---

## Bug Description

A clear and concise description of what the bug is.

## Steps to Reproduce

1. Go to '...'
2. Run command '...'
3. See error

## Expected Behavior

What you expected to happen.

## Actual Behavior

What actually happened.

## Environment

- **OS:** [e.g., macOS 14.1, Ubuntu 22.04, Windows 11]
- **ACM Version:** [e.g., 1.0.0]
- **Password Manager:** [e.g., 1Password CLI 2.x, Bitwarden CLI]
- **Go Version:** [if building from source]

## Logs

```
Paste relevant log output here
```

## Additional Context

Add any other context about the problem here (screenshots, config files, etc.).

## Possible Solution (Optional)

If you have ideas on how to fix this, please share!
EOF

# Feature Request Template
cat > .github/ISSUE_TEMPLATE/feature_request.md << 'EOF'
---
name: Feature Request
about: Suggest a new feature or enhancement
title: '[FEATURE] '
labels: 'enhancement, needs-triage'
assignees: ''
---

## Feature Summary

A clear and concise description of the feature you'd like to see.

## Use Case

Describe the problem this feature would solve or the use case it addresses.

## Proposed Solution

Describe how you envision this feature working.

## Alternatives Considered

Have you considered alternative solutions? If so, what and why did you choose this approach?

## Additional Context

Add any other context, mockups, or examples about the feature request here.

## Complexity Estimate (Optional)

- [ ] Small (< 1 day)
- [ ] Medium (1-3 days)
- [ ] Large (1-2 weeks)
- [ ] Very Large (2+ weeks)

## Related Issues

Are there any existing issues related to this? If so, link them here.
EOF

# Security Vulnerability Template
cat > .github/ISSUE_TEMPLATE/security_vulnerability.md << 'EOF'
---
name: Security Vulnerability (Public Template)
about: For non-sensitive security issues only
title: '[SECURITY] '
labels: 'security, needs-triage'
assignees: ''
---

⚠️ **WARNING:** Do NOT use this template for sensitive security vulnerabilities!

For sensitive security issues, please email: **security@acm.dev**

Use PGP encryption: https://acm-project.dev/pgp

---

## Security Issue Type

- [ ] Information disclosure (low severity)
- [ ] Denial of Service
- [ ] Other non-sensitive security concern

## Description

Describe the security issue.

## Impact

What is the potential impact of this issue?

## Mitigation

How can users protect themselves until a fix is released?
EOF

echo -e "${GREEN}✓ GitHub Issue templates created${NC}"

# Create Pull Request Template
echo -e "${BLUE}[6/10] Creating Pull Request template...${NC}"
cat > .github/PULL_REQUEST_TEMPLATE/pull_request_template.md << 'EOF'
## Description

[Brief description of the changes in this PR]

## Related Issue

Closes #[issue number]

## Type of Change

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📝 Documentation update
- [ ] 🎨 Code style update (formatting, renaming)
- [ ] ♻️ Code refactoring (no functional changes)
- [ ] ⚡ Performance improvement
- [ ] ✅ Test update
- [ ] 🔧 Build/configuration update

## Component

- [ ] Core Service
- [ ] CRS (Credential Remediation)
- [ ] ACVS (Compliance Validation)
- [ ] HIM Manager
- [ ] OpenTUI
- [ ] Tauri GUI
- [ ] Security
- [ ] Legal/Compliance
- [ ] Documentation
- [ ] Infrastructure

## Testing

Describe the testing you performed:

- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed
- [ ] E2E tests added/updated (if applicable)

**Testing Notes:**
```
[Describe manual testing steps, if any]
```

## Screenshots (if applicable)

[Add screenshots for UI changes]

## Checklist

- [ ] My code follows the project's style guidelines
- [ ] I have performed a self-review of my code
- [ ] I have commented my code, particularly in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing unit tests pass locally with my changes
- [ ] Any dependent changes have been merged and published
- [ ] I have updated CHANGELOG.md (if user-facing change)

## Additional Notes

[Any additional information that reviewers should know]
EOF
echo -e "${GREEN}✓ Pull Request template created${NC}"

# Create Discord bot setup script
echo -e "${BLUE}[7/10] Creating Discord bot setup script...${NC}"
cat > scripts/community/discord-bot.py << 'EOF'
#!/usr/bin/env python3
"""
ACM Discord Bot - Community Management and Integration

Features:
- Welcome new members
- GitHub issue/PR notifications
- Community call reminders
- Contributor recognition
- FAQ auto-responder
"""

import os
import discord
from discord.ext import commands, tasks
import aiohttp
from datetime import datetime, timedelta

# Configuration
DISCORD_TOKEN = os.getenv("DISCORD_BOT_TOKEN")
GITHUB_TOKEN = os.getenv("GITHUB_TOKEN")
GITHUB_REPO = "acm-project/acm"
WELCOME_CHANNEL_ID = int(os.getenv("WELCOME_CHANNEL_ID", "0"))
ANNOUNCEMENTS_CHANNEL_ID = int(os.getenv("ANNOUNCEMENTS_CHANNEL_ID", "0"))

intents = discord.Intents.default()
intents.message_content = True
intents.members = True

bot = commands.Bot(command_prefix="!", intents=intents)

@bot.event
async def on_ready():
    print(f"{bot.user} has connected to Discord!")
    check_community_call.start()
    check_github_activity.start()

@bot.event
async def on_member_join(member):
    """Welcome new members"""
    welcome_channel = bot.get_channel(WELCOME_CHANNEL_ID)
    if welcome_channel:
        embed = discord.Embed(
            title=f"Welcome to ACM, {member.name}! 👋",
            description=(
                f"We're glad you're here, {member.mention}!\n\n"
                "🔒 ACM is a local-first credential remediation tool.\n"
                "💬 Introduce yourself in #introductions\n"
                "📚 Check out our [documentation](https://docs.acm-project.dev)\n"
                "🐛 Found a bug? Open an [issue](https://github.com/acm-project/acm/issues)\n"
                "💡 Want to contribute? Read [CONTRIBUTING.md](https://github.com/acm-project/acm/blob/main/CONTRIBUTING.md)\n\n"
                "Questions? Ask in #help!"
            ),
            color=discord.Color.blue()
        )
        embed.set_thumbnail(url=member.avatar.url if member.avatar else member.default_avatar.url)
        await welcome_channel.send(embed=embed)

@bot.command(name="help-acm")
async def help_acm(ctx):
    """Show ACM help and resources"""
    embed = discord.Embed(
        title="ACM Resources & Help",
        description="Here's how to get started with ACM:",
        color=discord.Color.green()
    )
    embed.add_field(
        name="📖 Documentation",
        value="[docs.acm-project.dev](https://docs.acm-project.dev)",
        inline=False
    )
    embed.add_field(
        name="💻 Repository",
        value="[github.com/acm-project/acm](https://github.com/acm-project/acm)",
        inline=False
    )
    embed.add_field(
        name="🐛 Report a Bug",
        value="[Open an Issue](https://github.com/acm-project/acm/issues/new?template=bug_report.md)",
        inline=False
    )
    embed.add_field(
        name="💡 Feature Request",
        value="[Open an Issue](https://github.com/acm-project/acm/issues/new?template=feature_request.md)",
        inline=False
    )
    embed.add_field(
        name="🙋 Getting Help",
        value="Ask in #help or [GitHub Discussions](https://github.com/acm-project/acm/discussions)",
        inline=False
    )
    await ctx.send(embed=embed)

@bot.command(name="roadmap")
async def roadmap(ctx):
    """Show ACM development roadmap"""
    embed = discord.Embed(
        title="ACM Development Roadmap",
        description="Current development phases and progress:",
        color=discord.Color.purple()
    )
    embed.add_field(
        name="Phase I: MVP (Current)",
        value="✅ Core service\n✅ CRS module\n🔄 OpenTUI\n⏳ mTLS auth",
        inline=False
    )
    embed.add_field(
        name="Phase II: ACVS",
        value="⏳ Legal NLP\n⏳ Evidence chains\n⏳ Tauri GUI",
        inline=False
    )
    embed.add_field(
        name="📊 Project Board",
        value="[View on GitHub](https://github.com/orgs/acm-project/projects)",
        inline=False
    )
    await ctx.send(embed=embed)

@tasks.loop(hours=24)
async def check_community_call():
    """Remind about upcoming community call"""
    now = datetime.now()
    # Community call: First Thursday of each month, 5pm UTC
    first_thursday = datetime(now.year, now.month, 1)
    while first_thursday.weekday() != 3:  # Thursday = 3
        first_thursday += timedelta(days=1)
    
    first_thursday = first_thursday.replace(hour=17, minute=0, second=0)
    
    # Send reminder 24 hours before
    if first_thursday - timedelta(hours=24) <= now <= first_thursday - timedelta(hours=23):
        channel = bot.get_channel(ANNOUNCEMENTS_CHANNEL_ID)
        if channel:
            embed = discord.Embed(
                title="📞 Community Call Tomorrow!",
                description=(
                    f"Join us for the monthly ACM Community Call!\n\n"
                    f"**When:** Tomorrow at 5:00 PM UTC\n"
                    f"**Where:** [Zoom Link](https://zoom.us/j/acm-meeting)\n\n"
                    f"**Agenda:**\n"
                    f"• Project updates\n"
                    f"• RFC discussions\n"
                    f"• Open forum\n"
                    f"• Contributor spotlight\n\n"
                    f"All are welcome! 🎉"
                ),
                color=discord.Color.gold()
            )
            await channel.send(embed=embed)

@tasks.loop(hours=1)
async def check_github_activity():
    """Check for new GitHub issues, PRs, releases"""
    # This is a simplified example; production bot would use webhooks
    pass

if __name__ == "__main__":
    if not DISCORD_TOKEN:
        print("Error: DISCORD_BOT_TOKEN environment variable not set")
        exit(1)
    
    bot.run(DISCORD_TOKEN)
EOF
chmod +x scripts/community/discord-bot.py
echo -e "${GREEN}✓ Discord bot setup script created${NC}"

# Create contributor onboarding script
echo -e "${BLUE}[8/10] Creating contributor onboarding script...${NC}"
cat > scripts/community/onboard-contributor.sh << 'EOF'
#!/usr/bin/env bash
# Contributor Onboarding Script

set -euo pipefail

CONTRIBUTOR_NAME="$1"
GITHUB_USERNAME="$2"

echo "🎉 Welcome to ACM, $CONTRIBUTOR_NAME!"
echo ""
echo "Setting up your contributor profile..."

# Add to contributors list
if [ ! -f CONTRIBUTORS.md ]; then
    cat > CONTRIBUTORS.md << 'HEADER'
# Contributors

Thank you to everyone who has contributed to ACM!

## Core Maintainers

(To be filled)

## Contributors

HEADER
fi

# Add contributor
echo "- [@$GITHUB_USERNAME](https://github.com/$GITHUB_USERNAME) - $CONTRIBUTOR_NAME" >> CONTRIBUTORS.md
echo "✓ Added to CONTRIBUTORS.md"

# Create welcome issue
gh issue create \
    --title "Welcome @$GITHUB_USERNAME!" \
    --body "🎉 Welcome to the ACM project, @$GITHUB_USERNAME!

We're excited to have you as a contributor. Here are some resources to get you started:

## Getting Started
- [ ] Read [CONTRIBUTING.md](CONTRIBUTING.md)
- [ ] Set up your development environment
- [ ] Join our [Discord](https://discord.gg/acm)
- [ ] Introduce yourself in #introductions

## First Contribution Ideas
Check out issues labeled [good first issue](https://github.com/acm-project/acm/labels/good%20first%20issue)

## Questions?
- Ask in #help on Discord
- Post in [GitHub Discussions](https://github.com/acm-project/acm/discussions)
- Tag @core-team in issues

Looking forward to your contributions! 🚀" \
    --label "welcome"

echo "✓ Created welcome issue"

echo ""
echo "Next steps for $CONTRIBUTOR_NAME:"
echo "1. Check your GitHub notifications for welcome issue"
echo "2. Join Discord: https://discord.gg/acm"
echo "3. Introduce yourself in #introductions"
echo "4. Pick a 'good first issue' to work on"
echo ""
echo "Thank you for joining the ACM community! 🙏"
EOF
chmod +x scripts/community/onboard-contributor.sh
echo -e "${GREEN}✓ Contributor onboarding script created${NC}"

# Create GitHub Actions workflow for community automation
echo -e "${BLUE}[9/10] Creating GitHub Actions workflow...${NC}"
mkdir -p .github/workflows
cat > .github/workflows/community-automation.yml << 'EOF'
name: Community Automation

on:
  issues:
    types: [opened, labeled]
  pull_request:
    types: [opened, merged]
  schedule:
    - cron: '0 0 * * *'  # Daily at midnight UTC

jobs:
  welcome-new-contributor:
    name: Welcome New Contributors
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request' && github.event.action == 'opened'
    steps:
      - name: Check if first-time contributor
        uses: actions/github-script@v6
        with:
          script: |
            const author = context.payload.pull_request.user.login;
            const { data: prs } = await github.rest.pulls.list({
              owner: context.repo.owner,
              repo: context.repo.repo,
              state: 'all',
              creator: author
            });
            
            if (prs.length === 1) {
              // First PR!
              await github.rest.issues.createComment({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: context.payload.pull_request.number,
                body: `🎉 Welcome @${author}! This is your first Pull Request to ACM.\n\nThank you for your contribution! A maintainer will review your PR soon.\n\nWhile you wait:\n- Join our [Discord](https://discord.gg/acm)\n- Check out other [good first issues](https://github.com/${context.repo.owner}/${context.repo.repo}/labels/good%20first%20issue)\n\nQuestions? Ask in #help on Discord!`
              });
              
              // Add label
              await github.rest.issues.addLabels({
                owner: context.repo.owner,
                repo: context.repo.repo,
                issue_number: context.payload.pull_request.number,
                labels: ['first-time-contributor']
              });
            }

  celebrate-merged-pr:
    name: Celebrate Merged PR
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request' && github.event.action == 'merged'
    steps:
      - name: Thank contributor
        uses: actions/github-script@v6
        with:
          script: |
            const author = context.payload.pull_request.user.login;
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.payload.pull_request.number,
              body: `🎉 Congratulations @${author}, your PR has been merged!\n\nYour contribution will be included in the next release. Thank you for making ACM better! 🙏\n\nWant to contribute more? Check out [open issues](https://github.com/${context.repo.owner}/${context.repo.repo}/issues).`
            });

  label-new-issues:
    name: Auto-label New Issues
    runs-on: ubuntu-latest
    if: github.event_name == 'issues' && github.event.action == 'opened'
    steps:
      - name: Add needs-triage label
        uses: actions/github-script@v6
        with:
          script: |
            await github.rest.issues.addLabels({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.payload.issue.number,
              labels: ['needs-triage']
            });
EOF
echo -e "${GREEN}✓ GitHub Actions workflow created${NC}"

# Create README template
echo -e "${BLUE}[10/10] Creating README.md template...${NC}"
cat > README.md << 'EOF'
# ACM - Automated Compromise Mitigation

🔒 **Local-first credential breach response with zero-knowledge security**

[![CI](https://github.com/acm-project/acm/workflows/CI/badge.svg)](https://github.com/acm-project/acm/actions)
[![Security Audit](https://img.shields.io/badge/security-audited-brightgreen.svg)](docs/security/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Discord](https://img.shields.io/discord/YOUR_DISCORD_ID?label=discord)](https://discord.gg/acm)

---

## What is ACM?

ACM automatically detects and rotates compromised credentials following data breaches while maintaining strict zero-knowledge security principles.

### Key Features

- 🔐 **Zero-Knowledge Architecture** - Your master password never leaves your device
- 🏠 **Local-First** - All processing happens on your machine
- ⚖️ **ToS Compliance** - ACVS validates automation against website Terms of Service
- 🔍 **Tamper-Evident Audit Logs** - Cryptographically signed evidence chains
- 🖥️ **Dual Interface** - OpenTUI for developers, Tauri GUI for everyone
- 🔓 **Open Source** - Fully auditable, community-driven

---

## Quick Start

### Installation

**macOS:**
```bash
brew tap acm-project/acm
brew install acm
```

**Linux:**
```bash
# Debian/Ubuntu
sudo apt install acm

# Fedora
sudo dnf install acm
```

**Windows:**
```powershell
choco install acm
```

### Setup

```bash
# Initialize ACM and generate certificates
acm setup

# Start the service
acm service start

# Detect compromised credentials
acm detect

# Rotate a credential
acm rotate <credential-id>
```

---

## Documentation

- 📖 [User Guide](docs/guides/user-guide.md)
- 🏗️ [Architecture](docs/architecture/TAD.md)
- 🔒 [Security](docs/security/security-planning.md)
- ⚖️ [Legal Framework](docs/legal/legal-framework.md)
- 🤝 [Contributing](CONTRIBUTING.md)

---

## Project Status

**Current Phase:** Phase I (MVP) — In Development

| Component | Status |
|-----------|--------|
| Core Service | 🔄 In Progress |
| CRS Module | 🔄 In Progress |
| OpenTUI | ⏳ Planned |
| mTLS Auth | ⏳ Planned |
| ACVS | ⏳ Phase II |
| Tauri GUI | ⏳ Phase II |

[View Roadmap](https://github.com/orgs/acm-project/projects) | [Track Progress](https://github.com/acm-project/acm/issues)

---

## Community

### Get Involved

- 💬 [Discord](https://discord.gg/acm) - Real-time chat and support
- 🗨️ [GitHub Discussions](https://github.com/acm-project/acm/discussions) - Questions and ideas
- 📅 [Monthly Community Call](docs/community/events.md) - First Thursday, 5pm UTC
- 🐦 [Twitter/X](https://twitter.com/acm_project) - Updates and announcements

### Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code contribution workflow
- Development setup
- Coding standards
- PR process

**Good First Issues:** [View beginner-friendly tasks](https://github.com/acm-project/acm/labels/good%20first%20issue)

---

## Security

Found a security vulnerability? **Do not open a public issue.**

Email: **security@acm.dev** (PGP key: [link](https://acm-project.dev/pgp))

See [SECURITY.md](SECURITY.md) for our responsible disclosure policy.

---

## License

ACM is licensed under the [MIT License](LICENSE).

By using ACM, you agree to the [End User License Agreement (EULA)](docs/legal/EULA.md).

---

## Acknowledgments

ACM is inspired by:
- **1Password** and **Bitwarden** - Password manager security best practices
- **Local-First Software** - Zero-knowledge and privacy-first principles
- **Open-source security projects** - Transparency and community-driven development

Special thanks to our contributors! See [CONTRIBUTORS.md](CONTRIBUTORS.md).

---

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=acm-project/acm&type=Date)](https://star-history.com/#acm-project/acm&Date)

---

**Made with ❤️ by the ACM community**

[Website](https://acm-project.dev) • [Docs](https://docs.acm-project.dev) • [Discord](https://discord.gg/acm) • [Twitter](https://twitter.com/acm_project)
EOF
echo -e "${GREEN}✓ README.md template created${NC}"

# Copy everything to outputs
echo -e "${BLUE}Copying files to outputs...${NC}"
cp setup-github-project.sh /mnt/user-data/outputs/
cp -r .github /home/claude/github-templates/
cp CONTRIBUTING.md CODE_OF_CONDUCT.md README.md /home/claude/
cp scripts/community/* /home/claude/community-scripts/ 2>/dev/null || true

echo ""
echo -e "${GREEN}╔══════════════════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Community Building Setup Complete!                     ║${NC}"
echo -e "${GREEN}╚══════════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Created Files:${NC}"
echo "✓ CONTRIBUTING.md"
echo "✓ CODE_OF_CONDUCT.md"
echo "✓ README.md"
echo "✓ .github/ISSUE_TEMPLATE/ (bug, feature, security)"
echo "✓ .github/PULL_REQUEST_TEMPLATE/"
echo "✓ .github/workflows/community-automation.yml"
echo "✓ scripts/community/discord-bot.py"
echo "✓ scripts/community/onboard-contributor.sh"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Run ./setup-github-project.sh to create GitHub Project"
echo "2. Set up Discord server and configure bot"
echo "3. Customize templates for your specific needs"
echo "4. Set up GitHub webhooks for Discord notifications"
echo "5. Schedule first community call"
echo ""
echo -e "${YELLOW}Discord Bot Setup:${NC}"
echo "export DISCORD_BOT_TOKEN='your-token'"
echo "export WELCOME_CHANNEL_ID='channel-id'"
echo "export ANNOUNCEMENTS_CHANNEL_ID='channel-id'"
echo "python3 scripts/community/discord-bot.py"
echo ""
EOF
chmod +x /home/claude/community-building-setup.sh
echo -e "${GREEN}✓ Community building script created${NC}"

# Copy all files to outputs
cp /home/claude/acm-security-planning.md /mnt/user-data/outputs/
cp /home/claude/setup-github-project.sh /mnt/user-data/outputs/
cp /home/claude/community-building-setup.sh /mnt/user-data/outputs/

echo "All files created and ready!"
