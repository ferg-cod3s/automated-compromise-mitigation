# Community Building Scripts & Setup
# Automated Compromise Mitigation (ACM)

**Version:** 1.0  
**Date:** November 2025  
**Purpose:** Automated scripts and templates for community building and management

---

## 1. Community Infrastructure Setup

### 1.1 Setup Checklist

```bash
# Community Infrastructure Setup Checklist

✅ Communication Channels:
   - [ ] Create Discord server
   - [ ] Setup GitHub Discussions
   - [ ] Create mailing lists (announce@, security@, legal@)
   - [ ] Setup Twitter/X account (@acm_project)
   - [ ] Create project blog (Hugo or Jekyll)

✅ Documentation:
   - [ ] Create project website (acm-project.dev)
   - [ ] Setup documentation hosting (docs.acm-project.dev)
   - [ ] Add CONTRIBUTING.md
   - [ ] Add CODE_OF_CONDUCT.md
   - [ ] Add SECURITY.md
   - [ ] Create community guidelines

✅ Automation:
   - [ ] Setup welcome bot (Discord + GitHub)
   - [ ] Configure contributor recognition automation
   - [ ] Setup newsletter automation
   - [ ] Create community metrics dashboard
```

---

## 2. Discord Server Setup

### 2.1 Discord Server Structure

**Server Name:** ACM Project

**Category: 📢 ANNOUNCEMENTS**
- `#announcements` (Read-only, for maintainers)
- `#releases` (Auto-posted from GitHub via webhook)
- `#community-updates` (Monthly community call notes, contributor spotlights)

**Category: 💬 GENERAL**
- `#general` (General discussion)
- `#introductions` (New members introduce themselves)
- `#show-and-tell` (Share what you've built with ACM)
- `#random` (Off-topic chat)

**Category: 🛠️ DEVELOPMENT**
- `#development` (General dev discussion)
- `#crs-module` (Credential Remediation Service)
- `#acvs-module` (Compliance Validation)
- `#ui-ux` (OpenTUI and Tauri GUI)
- `#testing` (Testing, QA, bug reports)

**Category: 🔐 SECURITY & LEGAL**
- `#security` (Security discussions, not for vulnerability reports)
- `#legal-compliance` (Legal NLP, ToS analysis, EULA discussions)
- `#threat-modeling` (Security architecture discussions)

**Category: 📚 HELP & SUPPORT**
- `#help-desk` (User support questions)
- `#installation-setup` (Installation troubleshooting)
- `#password-managers` (Password manager CLI help)
- `#faq` (Frequently asked questions)

**Category: 🎯 WORKING GROUPS**
- `#wg-password-integration` (Password Manager Integration WG)
- `#wg-legal-nlp` (Legal NLP Model WG)
- `#wg-documentation` (Documentation WG)
- `#wg-design` (UX/Design WG)

**Category: 🤖 BOTS & LOGS**
- `#github-feed` (GitHub activity feed)
- `#bot-commands` (Bot command testing)

**Voice Channels:**
- `🎙️ Community Call` (Monthly community meetings)
- `🎙️ Dev Sync` (Ad-hoc development discussions)
- `🎙️ Working Group Rooms` (WG-specific voice channels)

### 2.2 Discord Roles

| Role | Color | Permissions | How to Get |
|------|-------|-------------|------------|
| **@Project Lead** | Red | Admin | Project founder |
| **@Core Maintainer** | Orange | Moderator | Appointed by Project Lead |
| **@Security Lead** | Dark Red | Moderator | Security-focused maintainer |
| **@Community Manager** | Blue | Moderator | Community management role |
| **@Contributor** | Green | Standard | Merge 1+ PR |
| **@Core Contributor** | Bright Green | Standard | Merge 10+ PRs |
| **@Working Group Lead** | Purple | Standard | Lead a working group |
| **@Beta Tester** | Yellow | Standard | Active in beta testing |
| **@Donor** | Gold | Standard | Financial supporter |
| **@Member** | Gray | Standard | Default role |

### 2.3 Discord Bot Configuration

**Bot:** ACM Helper Bot (custom bot using discord.js or discord.py)

**Commands:**

```
!help - Show available commands
!docs <topic> - Link to documentation
!issue <number> - Link to GitHub issue
!pr <number> - Link to GitHub pull request
!contribute - Show contribution guidelines
!security - Show security disclosure policy
!coc - Show Code of Conduct
!setup - Link to setup guide
!welcome @user - Welcome new user (moderators only)
!thank @user - Thank contributor (auto-grants @Contributor role)
```

**Auto-Moderation Rules:**

```yaml
# Spam Prevention
- Delete messages with 5+ identical emojis in a row
- Rate limit: Max 5 messages per 5 seconds per user
- Ban on 3+ spam warnings

# Prohibited Content
- Delete messages containing discord invite links (except moderators)
- Delete messages with NSFW content
- Flag messages with potential security vulnerabilities for review

# Helpful Auto-Responses
- "master password" keyword → Link to security best practices
- "install" or "setup" → Link to installation guide
- "bug" → Link to bug report template
```

---

## 3. GitHub Discussions Configuration

### 3.1 Discussion Categories

**Category: 📣 Announcements**
- Format: Announcement
- Permissions: Maintainers can create, all can comment
- Purpose: Official project announcements

**Category: 💡 Ideas**
- Format: Open-ended discussion
- Purpose: Feature ideas, brainstorming
- Auto-label: `enhancement` if converted to issue

**Category: ❓ Q&A**
- Format: Question/Answer
- Purpose: Technical questions, best practices
- Allows marking answers

**Category: 🎨 Show and Tell**
- Format: Open-ended discussion
- Purpose: Share integrations, tutorials, use cases

**Category: 🗣️ General**
- Format: Open-ended discussion
- Purpose: Anything not fitting other categories

**Category: 📊 Polls**
- Format: Poll
- Purpose: Community voting on features, decisions

**Category: 🔐 Security (Private)**
- Format: Private (maintainers only)
- Purpose: Internal security discussions

---

## 4. Automated Welcome Messages

### 4.1 GitHub Welcome Bot

**File:** `.github/workflows/welcome.yml`

```yaml
name: Welcome New Contributors

on:
  pull_request_target:
    types: [opened]
  issues:
    types: [opened]

jobs:
  welcome:
    runs-on: ubuntu-latest
    steps:
      - name: Welcome new contributor (PR)
        if: github.event_name == 'pull_request_target'
        uses: actions/github-script@v7
        with:
          script: |
            const creator = context.payload.pull_request.user.login;
            const repo = context.repo;
            
            // Check if this is user's first PR
            const { data: prs } = await github.rest.pulls.list({
              owner: repo.owner,
              repo: repo.repo,
              creator: creator,
              state: 'all'
            });
            
            if (prs.length === 1) {
              // First PR!
              await github.rest.issues.createComment({
                owner: repo.owner,
                repo: repo.repo,
                issue_number: context.payload.pull_request.number,
                body: `🎉 Welcome to ACM, @${creator}! 
                
Thank you for your first contribution! Here's what happens next:

1. **Automated Checks**: Our CI/CD pipeline will run tests and linters. Make sure all checks pass.
2. **Code Review**: A maintainer will review your changes within 48 hours.
3. **Iteration**: You may be asked to make changes. No worries - this is normal!
4. **Merge**: Once approved and checks pass, we'll merge your PR.

**Resources:**
- [Contributing Guide](https://github.com/${repo.owner}/${repo.repo}/blob/main/CONTRIBUTING.md)
- [Code Style Guide](https://github.com/${repo.owner}/${repo.repo}/wiki/Code-Style)
- [Discord Server](https://discord.gg/acm) - Join us for real-time help!

Thanks again for contributing! 🚀`
              });
            }
      
      - name: Welcome new contributor (Issue)
        if: github.event_name == 'issues'
        uses: actions/github-script@v7
        with:
          script: |
            const creator = context.payload.issue.user.login;
            const repo = context.repo;
            
            // Check if this is user's first issue
            const { data: issues } = await github.rest.issues.listForRepo({
              owner: repo.owner,
              repo: repo.repo,
              creator: creator,
              state: 'all'
            });
            
            if (issues.length === 1) {
              // First issue!
              await github.rest.issues.createComment({
                owner: repo.owner,
                repo: repo.repo,
                issue_number: context.payload.issue.number,
                body: `👋 Welcome to ACM, @${creator}!

Thanks for opening your first issue. A maintainer will review it within 48 hours.

**In the meantime:**
- Make sure you've filled out the issue template completely
- Check if this is a duplicate of an existing issue
- If this is a security vulnerability, please report it privately to security@acm.dev instead

**Need help?**
- [FAQ](https://github.com/${repo.owner}/${repo.repo}/wiki/FAQ)
- [Discord Support](https://discord.gg/acm)

Thank you for helping improve ACM! 🙏`
              });
            }
```

### 4.2 Discord Welcome Message

**Discord Bot Script:**

```javascript
// discord-welcome-bot.js

const Discord = require('discord.js');
const client = new Discord.Client({ intents: ['Guilds', 'GuildMembers', 'GuildMessages'] });

client.on('guildMemberAdd', async member => {
  const welcomeChannel = member.guild.channels.cache.find(ch => ch.name === 'introductions');
  
  if (!welcomeChannel) return;
  
  const welcomeEmbed = new Discord.EmbedBuilder()
    .setColor('#0099ff')
    .setTitle(`Welcome to ACM Project, ${member.user.username}! 🎉`)
    .setDescription(`We're excited to have you here!`)
    .addFields(
      { name: '📖 Start Here', value: 'Read our [Contributing Guide](https://github.com/acm-project/acm/blob/main/CONTRIBUTING.md)' },
      { name: '👋 Introduce Yourself', value: 'Tell us about yourself in this channel!' },
      { name: '🛠️ Find an Issue', value: 'Check out [Good First Issues](https://github.com/acm-project/acm/labels/good-first-issue)' },
      { name: '💬 Get Help', value: 'Ask questions in <#help-desk>' },
      { name: '📚 Documentation', value: '[docs.acm-project.dev](https://docs.acm-project.dev)' }
    )
    .setThumbnail(member.user.displayAvatarURL())
    .setTimestamp();
  
  await welcomeChannel.send({ embeds: [welcomeEmbed] });
  
  // Send DM to new member
  try {
    await member.send(`Hey ${member.user.username}! 👋

Welcome to the ACM Project community! We're building a local-first, zero-knowledge tool for automated credential breach response.

**Quick links to get started:**
🔗 GitHub: https://github.com/acm-project/acm
📖 Docs: https://docs.acm-project.dev
💬 Introduce yourself in #introductions
🛠️ Find issues to work on: https://github.com/acm-project/acm/labels/good-first-issue

**Community Guidelines:**
✅ Be respectful and inclusive
✅ Ask questions - we're here to help!
✅ Share your ideas and feedback
❌ No spam or self-promotion
❌ No security vulnerability disclosure in public channels (email security@acm.dev instead)

See you around! 🚀`);
  } catch (error) {
    console.log(`Could not send DM to ${member.user.tag}: ${error}`);
  }
});

client.login(process.env.DISCORD_BOT_TOKEN);
```

---

## 5. Contributor Recognition Automation

### 5.1 Monthly Contributor Spotlight

**Script:** `scripts/contributor-spotlight.sh`

```bash
#!/bin/bash
# Generate monthly contributor spotlight

MONTH=$(date -d "last month" +%Y-%m)
OUTPUT_FILE="content/blog/contributor-spotlight-$MONTH.md"

echo "Generating contributor spotlight for $MONTH..."

# Get top contributors by commits
TOP_COMMITTERS=$(gh api graphql -f query='
  query {
    repository(owner: "acm-project", name: "acm") {
      defaultBranchRef {
        target {
          ... on Commit {
            history(first: 100, since: "'$MONTH'-01T00:00:00Z", until: "'$MONTH'-31T23:59:59Z") {
              edges {
                node {
                  author {
                    user {
                      login
                      name
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
' | jq -r '.data.repository.defaultBranchRef.target.history.edges[].node.author.user.login' | sort | uniq -c | sort -rn | head -5)

# Get top PR contributors
TOP_PR_AUTHORS=$(gh pr list --repo acm-project/acm --state merged --search "merged:$MONTH" --json author --jq '[.[] | .author.login] | group_by(.) | map({user: .[0], count: length}) | sort_by(.count) | reverse | .[0:5]')

# Generate blog post
cat > $OUTPUT_FILE <<EOF
---
title: "Contributor Spotlight: $(date -d "$MONTH-01" +"%B %Y")"
date: $(date +%Y-%m-%d)
author: Community Team
tags: [community, contributors]
---

# 🌟 Contributor Spotlight: $(date -d "$MONTH-01" +"%B %Y")

Thank you to everyone who contributed to ACM this month! Here are some of our top contributors:

## Top Committers

$TOP_COMMITTERS

## Top PR Contributors

$TOP_PR_AUTHORS

## Special Recognition

[Manually add special recognition here]

---

**Want to be featured next month?** Check out our [Good First Issues](https://github.com/acm-project/acm/labels/good-first-issue) and start contributing!

EOF

echo "Spotlight generated: $OUTPUT_FILE"
echo "Please review and add special recognition sections before publishing."
```

### 5.2 GitHub Actions: Auto-Thank Contributors

**File:** `.github/workflows/thank-contributors.yml`

```yaml
name: Thank Contributors

on:
  pull_request:
    types: [closed]

jobs:
  thank:
    if: github.event.pull_request.merged == true
    runs-on: ubuntu-latest
    
    steps:
      - name: Thank contributor for merged PR
        uses: actions/github-script@v7
        with:
          script: |
            const contributor = context.payload.pull_request.user.login;
            const prNumber = context.payload.pull_request.number;
            const prTitle = context.payload.pull_request.title;
            
            // Count total merged PRs by this contributor
            const { data: prs } = await github.rest.pulls.list({
              owner: context.repo.owner,
              repo: context.repo.repo,
              state: 'closed',
              creator: contributor
            });
            
            const mergedPRs = prs.filter(pr => pr.merged_at !== null);
            const prCount = mergedPRs.length;
            
            let message = `🎉 Thank you @${contributor} for your contribution! `;
            
            if (prCount === 1) {
              message += `This is your first merged PR - welcome to the ACM contributor community! 🚀

We're excited to have you as part of the team. Here's what you've unlocked:
- ✅ @Contributor role on Discord (if you've joined)
- ✅ Your name in the CONTRIBUTORS.md file
- ✅ Eligibility for Contributor of the Month

Keep up the great work!`;
            } else if (prCount === 10) {
              message += `This is your 10th merged PR! 🏆

You've earned the @Core Contributor badge! Your sustained contributions are making ACM better for everyone.

Next milestone: 25 PRs for Hall of Fame recognition!`;
            } else if (prCount === 25) {
              message += `🌟 AMAZING! This is your 25th merged PR! 🌟

You're now in the ACM Hall of Fame! Your dedication to the project is truly appreciated. Thank you for being such an integral part of our community.`;
            } else {
              message += `PR #${prNumber} has been merged. You've now contributed ${prCount} merged PRs to ACM!`;
            }
            
            await github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: prNumber,
              body: message
            });
```

---

## 6. Community Metrics Dashboard

### 6.1 Metrics Collection Script

**Script:** `scripts/collect-metrics.py`

```python
#!/usr/bin/env python3
"""
Collect community metrics for ACM project
"""

import os
import json
import requests
from datetime import datetime, timedelta

GITHUB_TOKEN = os.getenv('GITHUB_TOKEN')
REPO = 'acm-project/acm'
HEADERS = {'Authorization': f'token {GITHUB_TOKEN}'}

def github_api(endpoint):
    """Make GitHub API request"""
    url = f'https://api.github.com/{endpoint}'
    response = requests.get(url, headers=HEADERS)
    return response.json()

def collect_metrics():
    """Collect all metrics"""
    
    # Repository stats
    repo_data = github_api(f'repos/{REPO}')
    stars = repo_data.get('stargazers_count', 0)
    forks = repo_data.get('forks_count', 0)
    watchers = repo_data.get('subscribers_count', 0)
    
    # Contributors (30 days)
    since = (datetime.now() - timedelta(days=30)).isoformat()
    commits = github_api(f'repos/{REPO}/commits?since={since}&per_page=100')
    contributors = set(c['commit']['author']['name'] for c in commits if c.get('commit'))
    
    # Issues
    issues_open = github_api(f'repos/{REPO}/issues?state=open&per_page=1')
    issues_closed = github_api(f'repos/{REPO}/issues?state=closed&per_page=100')
    
    # Pull requests
    prs_open = github_api(f'repos/{REPO}/pulls?state=open&per_page=1')
    prs_merged = github_api(f'repos/{REPO}/pulls?state=closed&per_page=100')
    
    # Calculate metrics
    metrics = {
        'date': datetime.now().isoformat(),
        'github': {
            'stars': stars,
            'forks': forks,
            'watchers': watchers,
            'contributors_30d': len(contributors),
            'issues': {
                'open': len(issues_open),
                'closed_30d': len([i for i in issues_closed if is_within_30_days(i['closed_at'])])
            },
            'pull_requests': {
                'open': len(prs_open),
                'merged_30d': len([pr for pr in prs_merged if pr.get('merged_at') and is_within_30_days(pr['merged_at'])])
            }
        }
    }
    
    return metrics

def is_within_30_days(date_str):
    """Check if date is within last 30 days"""
    if not date_str:
        return False
    date = datetime.fromisoformat(date_str.replace('Z', '+00:00'))
    return (datetime.now(date.tzinfo) - date).days <= 30

def save_metrics(metrics):
    """Save metrics to file"""
    output_file = 'metrics/community-metrics.json'
    os.makedirs('metrics', exist_ok=True)
    
    # Append to historical metrics
    history = []
    if os.path.exists(output_file):
        with open(output_file, 'r') as f:
            history = json.load(f)
    
    history.append(metrics)
    
    with open(output_file, 'w') as f:
        json.dump(history, f, indent=2)
    
    print(f"Metrics saved to {output_file}")
    print(json.dumps(metrics, indent=2))

if __name__ == '__main__':
    metrics = collect_metrics()
    save_metrics(metrics)
```

### 6.2 Metrics Dashboard (HTML)

**File:** `website/metrics.html`

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>ACM Community Metrics</title>
    <script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
            background: #f5f5f5;
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin: 20px 0;
        }
        .metric-card {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .metric-card h3 {
            margin: 0 0 10px 0;
            color: #555;
            font-size: 14px;
            text-transform: uppercase;
        }
        .metric-value {
            font-size: 36px;
            font-weight: bold;
            color: #0066cc;
        }
        .metric-change {
            font-size: 14px;
            color: #28a745;
        }
        .chart-container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            margin: 20px 0;
        }
    </style>
</head>
<body>
    <h1>📊 ACM Community Metrics</h1>
    <p>Updated: <span id="last-updated"></span></p>
    
    <div class="metrics-grid">
        <div class="metric-card">
            <h3>⭐ GitHub Stars</h3>
            <div class="metric-value" id="stars">-</div>
            <div class="metric-change" id="stars-change">-</div>
        </div>
        
        <div class="metric-card">
            <h3>👥 Contributors (30d)</h3>
            <div class="metric-value" id="contributors">-</div>
            <div class="metric-change" id="contributors-change">-</div>
        </div>
        
        <div class="metric-card">
            <h3>🔧 Open Issues</h3>
            <div class="metric-value" id="issues-open">-</div>
            <div class="metric-change" id="issues-change">-</div>
        </div>
        
        <div class="metric-card">
            <h3>✅ Merged PRs (30d)</h3>
            <div class="metric-value" id="prs-merged">-</div>
            <div class="metric-change" id="prs-change">-</div>
        </div>
    </div>
    
    <div class="chart-container">
        <canvas id="contributorsChart"></canvas>
    </div>
    
    <div class="chart-container">
        <canvas id="activityChart"></canvas>
    </div>
    
    <script>
        // Load metrics data
        fetch('metrics/community-metrics.json')
            .then(response => response.json())
            .then(data => {
                const latest = data[data.length - 1];
                const previous = data[data.length - 2] || latest;
                
                // Update metric cards
                document.getElementById('last-updated').textContent = new Date(latest.date).toLocaleDateString();
                document.getElementById('stars').textContent = latest.github.stars;
                document.getElementById('contributors').textContent = latest.github.contributors_30d;
                document.getElementById('issues-open').textContent = latest.github.issues.open;
                document.getElementById('prs-merged').textContent = latest.github.pull_requests.merged_30d;
                
                // Calculate changes
                const starsChange = latest.github.stars - previous.github.stars;
                document.getElementById('stars-change').textContent = `+${starsChange} this period`;
                
                // Render charts
                renderContributorsChart(data);
                renderActivityChart(data);
            });
        
        function renderContributorsChart(data) {
            const ctx = document.getElementById('contributorsChart').getContext('2d');
            new Chart(ctx, {
                type: 'line',
                data: {
                    labels: data.map(d => new Date(d.date).toLocaleDateString()),
                    datasets: [{
                        label: 'Active Contributors (30d)',
                        data: data.map(d => d.github.contributors_30d),
                        borderColor: '#0066cc',
                        tension: 0.1
                    }]
                },
                options: {
                    responsive: true,
                    plugins: {
                        title: {
                            display: true,
                            text: 'Active Contributors Over Time'
                        }
                    }
                }
            });
        }
        
        function renderActivityChart(data) {
            const ctx = document.getElementById('activityChart').getContext('2d');
            new Chart(ctx, {
                type: 'bar',
                data: {
                    labels: data.map(d => new Date(d.date).toLocaleDateString()),
                    datasets: [
                        {
                            label: 'Issues Closed',
                            data: data.map(d => d.github.issues.closed_30d),
                            backgroundColor: '#28a745'
                        },
                        {
                            label: 'PRs Merged',
                            data: data.map(d => d.github.pull_requests.merged_30d),
                            backgroundColor: '#6f42c1'
                        }
                    ]
                },
                options: {
                    responsive: true,
                    plugins: {
                        title: {
                            display: true,
                            text: 'Project Activity (30 Days)'
                        }
                    }
                }
            });
        }
    </script>
</body>
</html>
```

---

## 7. Newsletter Automation

### 7.1 Monthly Newsletter Template

**File:** `templates/newsletter-template.md`

```markdown
# ACM Project Newsletter - {MONTH} {YEAR}

Welcome to the ACM Project newsletter! Here's what happened this month:

## 🚀 Highlights

{HIGHLIGHTS}

## 📊 Project Stats

- **GitHub Stars:** {STARS} (+{STARS_CHANGE} this month)
- **Active Contributors:** {CONTRIBUTORS}
- **Issues Closed:** {ISSUES_CLOSED}
- **Pull Requests Merged:** {PRS_MERGED}

## 🎉 New Contributors

Welcome to our new contributors this month:
{NEW_CONTRIBUTORS_LIST}

## 🌟 Contributor of the Month

{CONTRIBUTOR_SPOTLIGHT}

## 📢 Announcements

{ANNOUNCEMENTS}

## 🔧 Development Updates

### Phase I Progress: {PHASE_1_PROGRESS}%

{PHASE_1_UPDATES}

### Phase II Planning

{PHASE_2_UPDATES}

## 📚 New Documentation

{NEW_DOCS_LIST}

## 🗓️ Upcoming Events

- **Community Call:** {NEXT_COMMUNITY_CALL_DATE}
- **Contributor Summit:** {NEXT_SUMMIT_DATE}

## 💬 From the Community

{COMMUNITY_HIGHLIGHTS}

## 🔗 Quick Links

- [GitHub](https://github.com/acm-project/acm)
- [Documentation](https://docs.acm-project.dev)
- [Discord](https://discord.gg/acm)
- [Twitter](https://twitter.com/acm_project)

---

**Want to contribute?** Check out our [Good First Issues](https://github.com/acm-project/acm/labels/good-first-issue)!

**Have feedback?** Reply to this email or join us on Discord.

---

You're receiving this because you subscribed to ACM Project updates.
[Unsubscribe](https://acm-project.dev/unsubscribe?email={EMAIL})
```

### 7.2 Newsletter Generation Script

**Script:** `scripts/generate-newsletter.sh`

```bash
#!/bin/bash
# Generate monthly newsletter

MONTH=$(date +%B)
YEAR=$(date +%Y)
TEMPLATE="templates/newsletter-template.md"
OUTPUT="newsletters/newsletter-$(date +%Y-%m).md"

# Collect stats
STARS=$(gh api repos/acm-project/acm --jq '.stargazers_count')
STARS_LAST_MONTH=$(cat metrics/stars-last-month.txt 2>/dev/null || echo "0")
STARS_CHANGE=$((STARS - STARS_LAST_MONTH))

# Generate newsletter
cp $TEMPLATE $OUTPUT

# Replace placeholders
sed -i "s/{MONTH}/$MONTH/g" $OUTPUT
sed -i "s/{YEAR}/$YEAR/g" $OUTPUT
sed -i "s/{STARS}/$STARS/g" $OUTPUT
sed -i "s/{STARS_CHANGE}/$STARS_CHANGE/g" $OUTPUT

echo "Newsletter draft generated: $OUTPUT"
echo "Please fill in the following sections manually:"
echo "- {HIGHLIGHTS}"
echo "- {CONTRIBUTOR_SPOTLIGHT}"
echo "- {ANNOUNCEMENTS}"
echo ""
echo "After editing, send with: scripts/send-newsletter.sh $OUTPUT"
```

---

## 8. Community Call Automation

### 8.1 Community Call Reminder

**Script:** `scripts/community-call-reminder.sh`

```bash
#!/bin/bash
# Send community call reminders

CALL_DATE="First Thursday of next month, 5:00 PM UTC"
ZOOM_LINK="https://zoom.us/j/acm-community-call"

# Post to Discord
curl -X POST $DISCORD_WEBHOOK_URL \
  -H "Content-Type: application/json" \
  -d '{
    "embeds": [{
      "title": "📅 Community Call Reminder",
      "description": "Join us for our monthly community call!",
      "color": 3447003,
      "fields": [
        {
          "name": "📆 When",
          "value": "'"$CALL_DATE"'",
          "inline": false
        },
        {
          "name": "🔗 Join Link",
          "value": "'"$ZOOM_LINK"'",
          "inline": false
        },
        {
          "name": "📋 Agenda",
          "value": "[View agenda on GitHub Discussions](https://github.com/acm-project/acm/discussions)",
          "inline": false
        }
      ],
      "footer": {
        "text": "Add your topics to the agenda! React with 👍 if you plan to attend."
      }
    }]
  }'

# Post to GitHub Discussions
gh api repos/acm-project/acm/discussions \
  -X POST \
  -F title="Community Call - $CALL_DATE" \
  -F body="## 🎙️ Monthly Community Call

**Date:** $CALL_DATE  
**Link:** $ZOOM_LINK

### Agenda
1. Welcome and Introductions (5 min)
2. Project Updates (15 min)
3. RFC Discussions (20 min)
4. Open Forum (15 min)
5. Contributor Spotlight (5 min)

### Add Your Topics
Comment below to add discussion topics!

### Can't attend?
The call will be recorded and posted to YouTube."

echo "Community call reminders sent!"
```

### 8.2 Post-Call Summary

**Template:** `templates/community-call-summary.md`

```markdown
# Community Call Summary - {DATE}

**Attendees:** {ATTENDEE_COUNT}  
**Recording:** [YouTube Link]({YOUTUBE_URL})

## Key Discussions

### Project Updates
{PROJECT_UPDATES_SUMMARY}

### RFC Discussions
{RFC_DISCUSSIONS_SUMMARY}

### Open Forum
{OPEN_FORUM_SUMMARY}

### Contributor Spotlight
{CONTRIBUTOR_SPOTLIGHT_SUMMARY}

## Action Items

- [ ] {ACTION_ITEM_1} - @{ASSIGNEE_1}
- [ ] {ACTION_ITEM_2} - @{ASSIGNEE_2}
- [ ] {ACTION_ITEM_3} - @{ASSIGNEE_3}

## Next Call

**Date:** {NEXT_CALL_DATE}  
**Proposed Agenda Topics:**
- {TOPIC_1}
- {TOPIC_2}

---

**Add topics for next call:** Comment on this discussion

**Couldn't attend?** Watch the [recording]({YOUTUBE_URL})
```

---

## 9. Onboarding Checklist

### 9.1 New Contributor Onboarding

**File:** `docs/onboarding-checklist.md`

```markdown
# New Contributor Onboarding Checklist

Welcome to ACM! Here's your roadmap to becoming an active contributor.

## Getting Started (Week 1)

- [ ] ⭐ Star the [GitHub repository](https://github.com/acm-project/acm)
- [ ] 👁️ Watch the repository for notifications
- [ ] 📖 Read the [README](https://github.com/acm-project/acm/blob/main/README.md)
- [ ] 💬 Join [Discord](https://discord.gg/acm) and introduce yourself in #introductions
- [ ] 📚 Read [Contributing Guide](https://github.com/acm-project/acm/blob/main/CONTRIBUTING.md)
- [ ] 📝 Read [Code of Conduct](https://github.com/acm-project/acm/blob/main/CODE_OF_CONDUCT.md)

## Understanding the Project (Week 2)

- [ ] 🎯 Read [Product Requirements Document](docs/acm-prd.md)
- [ ] 🏗️ Read [Technical Architecture Document](docs/acm-tad.md)
- [ ] 🔐 Review [Threat Model](docs/acm-threat-model.md)
- [ ] 📊 Check [Project Roadmap](https://github.com/orgs/acm-project/projects/1)
- [ ] 🗂️ Browse [open issues](https://github.com/acm-project/acm/issues)

## Development Setup (Week 2-3)

- [ ] 🛠️ Follow [Development Setup Guide](docs/dev-setup.md)
- [ ] ✅ Run tests: `make test`
- [ ] 📦 Build the project: `make build`
- [ ] 🧪 Try running ACM locally

## First Contribution (Week 3-4)

- [ ] 🔍 Find a [Good First Issue](https://github.com/acm-project/acm/labels/good-first-issue)
- [ ] 💬 Comment on issue to claim it
- [ ] 🌿 Create a feature branch
- [ ] 💻 Write code following [Style Guide](docs/code-style.md)
- [ ] ✅ Add tests (>80% coverage)
- [ ] 📝 Update documentation
- [ ] 🔨 Create pull request
- [ ] 🎉 Celebrate your first PR!

## Becoming a Regular Contributor

- [ ] 🎙️ Attend monthly [Community Call](https://github.com/acm-project/acm/discussions)
- [ ] 🔄 Contribute 3+ pull requests
- [ ] 👥 Help others in #help-desk on Discord
- [ ] 📚 Improve documentation
- [ ] 🐛 Report bugs and suggest features
- [ ] ⚙️ Consider joining a [Working Group](docs/working-groups.md)

## Advanced Engagement

- [ ] 💬 Participate in RFC discussions
- [ ] 🔐 Report security vulnerabilities responsibly
- [ ] 🎓 Mentor new contributors
- [ ] 🏆 Aim for Contributor of the Month
- [ ] 🌟 Consider becoming a Core Contributor (10+ PRs)

---

**Questions?** Ask in #help-desk on Discord or open a [GitHub Discussion](https://github.com/acm-project/acm/discussions).
```

---

## 10. Implementation Checklist

```bash
# Community Building Setup Checklist

## Communication Channels
- [ ] Create Discord server with proper structure
- [ ] Configure Discord bot (ACM Helper Bot)
- [ ] Setup GitHub Discussions with categories
- [ ] Create mailing lists (announce@, security@, legal@)
- [ ] Register Twitter/X account (@acm_project)
- [ ] Setup project blog

## Automation
- [ ] Deploy GitHub Actions for welcome messages
- [ ] Deploy Discord welcome bot
- [ ] Setup contributor recognition automation
- [ ] Configure metrics collection (weekly cron)
- [ ] Setup newsletter generation scripts

## Documentation
- [ ] Create CONTRIBUTING.md
- [ ] Create CODE_OF_CONDUCT.md
- [ ] Create SECURITY.md
- [ ] Create onboarding checklist
- [ ] Create community guidelines

## Ongoing Tasks
- [ ] Schedule first community call
- [ ] Plan contributor summit (quarterly)
- [ ] Generate monthly newsletter
- [ ] Update metrics dashboard weekly
- [ ] Recognize contributors monthly
```

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-11-13 | Claude (AI Assistant) | Complete community building scripts and automation suite |

---

**Status:** Ready for Implementation  
**Next Steps:** Setup Discord server, configure bots, deploy GitHub Actions
