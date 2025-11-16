# Open-Source Governance & Roadmap
# Automated Compromise Mitigation (ACM)

**Version:** 1.0  
**Date:** November 2025  
**Status:** Draft  
**Document Type:** Governance Framework and Project Roadmap

---

## 1. Executive Summary

### 1.1 Purpose

This document establishes the governance framework for the Automated Compromise Mitigation (ACM) open-source project and defines the strategic roadmap for development from inception through maturity.

### 1.2 Governance Philosophy

ACM is governed by principles of:

- **Transparency**: All decisions, discussions, and code publicly visible
- **Meritocracy**: Contributions and expertise determine influence, not seniority
- **Consensus-Building**: Major decisions require broad community agreement
- **Benevolent Leadership**: Core maintainers guide direction while respecting community input
- **Inclusivity**: Welcoming contributors of all backgrounds and skill levels

### 1.3 Project Lifecycle Stage

**Current Stage:** Pre-Release (Planning Phase)  
**Target Stage:** Active Development → Stable Release → Mature Project

---

## 2. Project Vision and Mission

### 2.1 Vision Statement

**"Make secure credential breach response accessible, transparent, and automated for every user while maintaining zero-knowledge privacy and legal compliance."**

### 2.2 Mission

The ACM project exists to:

1. Provide a **local-first, zero-knowledge** credential remediation tool that operates without cloud dependencies
2. Demonstrate **good-faith ToS compliance** through automated validation (ACVS)
3. Maintain **radical transparency** via open-source development and security audits
4. Foster a **healthy, inclusive community** of contributors and users
5. Protect **project contributors** from legal liability through clear EULA and indemnification

### 2.3 Core Values

| Value | Description | How We Uphold It |
|-------|-------------|------------------|
| **Privacy First** | User data never leaves their device | Zero external API calls; local-first architecture |
| **Security by Design** | Security is not optional or added later | Threat modeling, security audits, responsible disclosure |
| **Legal Responsibility** | Users and project both protected | ACVS validates ToS; EULA transfers user liability |
| **Community Driven** | Major decisions require community input | RFC process for significant changes |
| **Transparency** | Open development, clear communication | Public roadmap, GitHub Discussions, monthly community calls |
| **Accessibility** | Usable by technical and non-technical users | TUI for power users, GUI for accessibility |

---

## 3. Governance Structure

### 3.1 Governance Model

ACM employs a **Benevolent Dictator Lite (BDL)** governance model with strong community input:

- **Project Lead** (initial maintainer) has final decision-making authority on critical issues
- **Core Maintainers** (3-5 individuals) collaborate on day-to-day decisions
- **Steering Committee** (formed at 50+ contributors) provides strategic guidance
- **Community** has voice via RFC process and GitHub Discussions

**Transition Plan**: If project grows beyond 100+ active contributors, transition to **Technical Steering Committee (TSC)** model with elected representatives.

### 3.2 Organizational Structure

```
┌─────────────────────────────────────────────────────────────┐
│                      ACM Community                          │
│  (All contributors, users, and community members)           │
└──────────────────────┬──────────────────────────────────────┘
                       │
        ┌──────────────┴──────────────┐
        │                             │
┌───────▼───────┐            ┌────────▼────────┐
│ Project Lead  │            │ Core Maintainers │
│               │            │  (3-5 people)    │
│ - Final       │            │                  │
│   decisions   │◄───────────┤ - Code review    │
│ - Vision      │            │ - Day-to-day     │
│ - Community   │            │   decisions      │
│   health      │            │ - Releases       │
└───────┬───────┘            └────────┬─────────┘
        │                             │
        │     ┌───────────────────────┴────────────────────┐
        │     │                                            │
        │ ┌───▼─────────┐  ┌──────────────┐  ┌──────────▼──────┐
        │ │  Security   │  │    Legal     │  │   Community     │
        │ │    Lead     │  │   Review     │  │    Manager      │
        │ │             │  │  Committee   │  │                 │
        │ │ - Audits    │  │              │  │ - Discord/forum │
        │ │ - CVEs      │  │ - EULA       │  │ - Events        │
        │ │ - Vulns     │  │ - ACVS       │  │ - Onboarding    │
        │ └─────────────┘  │   legal      │  └─────────────────┘
        │                  └──────────────┘
        │
        │ ┌────────────────────────────────────────────────────┐
        └─┤          Specialized Working Groups                 │
          │                                                      │
          │ - Password Manager Integration WG                   │
          │ - Legal NLP Model WG                                │
          │ - Documentation WG                                  │
          │ - UX/Design WG                                      │
          └──────────────────────────────────────────────────────┘
```

### 3.3 Roles and Responsibilities

#### 3.3.1 Project Lead

**Current:** [To Be Determined]  
**Term:** Indefinite (can step down voluntarily)

**Responsibilities:**
- Set overall project vision and strategic direction
- Make final decisions on contentious issues
- Represent project externally (conferences, press, partnerships)
- Maintain community health and culture
- Appoint Core Maintainers
- Ensure legal and security standards upheld

**Authority:**
- Final say on technical architecture decisions
- Veto power on proposals (rarely used)
- Appoint/remove Core Maintainers (with community consultation)

---

#### 3.3.2 Core Maintainers

**Current:** [To Be Determined — Target: 3-5 initial maintainers]  
**Term:** Indefinite (subject to removal for inactivity or misconduct)

**Selection Criteria:**
- Demonstrated sustained contributions (6+ months)
- Deep technical expertise in relevant areas (Go, security, legal, etc.)
- Excellent communication and collaboration skills
- Commitment to project values and Code of Conduct
- Available for weekly maintainer syncs

**Responsibilities:**
- Review and merge pull requests
- Triage GitHub Issues
- Release management (versioning, changelogs, distribution)
- Mentor new contributors
- Participate in RFC discussions
- Maintain code quality standards

**Authority:**
- Approve/reject pull requests (require 2+ approvals for major changes)
- Tag releases
- Grant repository write access to trusted contributors

**Expectations:**
- Minimum 5 hours/week contribution (flexible based on availability)
- Respond to reviews within 48 hours (or delegate)
- Attend monthly community calls (or send delegate)

---

#### 3.3.3 Security Lead

**Current:** [To Be Determined]  
**Term:** 1 year (renewable)

**Responsibilities:**
- Coordinate security audits and penetration testing
- Triage security vulnerability reports (security@acm.dev)
- Maintain security advisory process (CVEs, GitHub Security)
- Lead incident response for P0/P1 security incidents
- Review security-critical code changes
- Maintain threat model and security documentation

**Authority:**
- Emergency authority to disable features if critical vulnerability discovered
- Direct communication with security researchers
- Recommend security-related RFCs

---

#### 3.3.4 Legal Review Committee

**Current:** [To Be Formed — Target: 3-5 members including 1 licensed attorney]  
**Term:** 2 years (staggered terms)

**Selection Criteria:**
- Legal or compliance expertise (at least 1 licensed attorney)
- Understanding of open-source licensing
- Commitment to project's legal compliance mission

**Responsibilities:**
- Review EULA and legal documentation
- Provide guidance on Legal NLP model accuracy (not legal advice to users)
- Monitor legal developments (case law, regulations)
- Advise on incident response for legal threats
- Quarterly ACVS model accuracy review

**Authority:**
- Recommend EULA changes (final decision: Project Lead + Core Maintainers)
- Flag legal risks requiring external counsel
- Request temporary disabling of ACVS features pending legal review

**Disclaimer:** Legal Review Committee provides **guidance**, not legal advice. All members must clearly disclaim attorney-client relationship in public communications.

---

#### 3.3.5 Community Manager

**Current:** [To Be Determined]  
**Term:** 1 year (renewable)

**Responsibilities:**
- Moderate Discord, GitHub Discussions, and other community spaces
- Enforce Code of Conduct
- Organize community events (monthly calls, hackathons, etc.)
- Onboard new contributors
- Maintain contributor documentation (CONTRIBUTING.md, onboarding guides)
- Track community sentiment and report to Core Team

**Authority:**
- Issue warnings/temporary bans for Code of Conduct violations (permanent bans require Core Maintainer approval)
- Create and manage community spaces (Discord channels, GitHub Projects)
- Recognize and celebrate contributors (monthly spotlight, contributor badges)

---

### 3.4 Working Groups

**Purpose:** Focus on specific areas requiring sustained, coordinated effort.

**Current Working Groups:** (To be formed as needed)

1. **Password Manager Integration WG**
   - Maintain integrations with 1Password, Bitwarden, LastPass, etc.
   - Test compatibility with new CLI versions
   - Document integration patterns for new password managers

2. **Legal NLP Model WG**
   - Train and evaluate Legal NLP models for ToS analysis
   - Curate training data (ToS corpus)
   - Quarterly accuracy review and model updates

3. **Documentation WG**
   - Maintain user guides, API documentation, and tutorials
   - Create video tutorials and screencasts
   - Translate documentation to non-English languages (future)

4. **UX/Design WG**
   - Design and implement Tauri GUI
   - Improve OpenTUI user experience
   - Conduct user research and usability testing

**How to Form a Working Group:**
1. Propose WG via GitHub Discussion with clear scope and objectives
2. Recruit 3+ interested contributors
3. Assign a WG Lead (responsible for coordination and reporting)
4. Get Core Maintainer approval
5. WG provides monthly updates in community calls

---

## 4. Decision-Making Process

### 4.1 Decision-Making Framework

| Decision Type | Process | Authority | Example |
|---------------|---------|-----------|---------|
| **Trivial** | Single maintainer approval | Any Core Maintainer | Fix typo, update dependency version |
| **Minor** | 2+ maintainer approvals | Core Maintainers | Add new CLI flag, improve error message |
| **Major** | RFC process (1-2 week discussion) | Core Maintainers + community input | New feature, breaking API change |
| **Critical** | Extended RFC (3-4 weeks) + Project Lead approval | Project Lead + Core Maintainers + community | EULA change, governance change, major security decision |

### 4.2 RFC (Request for Comments) Process

**When to Use RFC:**
- Adding new major feature (e.g., new password manager support)
- Breaking changes to public APIs or CLI
- Architectural decisions with long-term impact
- Changes to EULA, licensing, or legal framework
- Governance or process changes

**RFC Workflow:**

```
1. DRAFT
   ├─ Author creates RFC document (markdown)
   ├─ Submit as Pull Request to acm-rfcs repository
   └─ Label: rfc-draft

2. DISCUSSION (1-4 weeks, depending on complexity)
   ├─ Community provides feedback in PR comments
   ├─ Author addresses concerns, updates RFC
   ├─ Core Maintainers participate in discussion
   └─ Label: rfc-in-discussion

3. FINAL COMMENT PERIOD (FCP)
   ├─ Core Maintainer initiates FCP (1 week for major, 2 weeks for critical)
   ├─ Label: rfc-fcp
   ├─ Last chance for objections
   └─ If no blocking concerns, proceed to decision

4. DECISION
   ├─ Core Maintainers vote (majority required)
   ├─ Project Lead has tiebreaker vote (and veto for critical RFCs)
   └─ Label: rfc-accepted OR rfc-rejected

5. IMPLEMENTATION (if accepted)
   ├─ RFC merged into acm-rfcs repository
   ├─ Implementation tracked via GitHub Issues/Projects
   └─ RFC number referenced in commits/PRs
```

**RFC Template:**

```markdown
# RFC-XXXX: [Title]

- Start Date: YYYY-MM-DD
- RFC PR: [acm-rfcs#XXX]
- Tracking Issue: [acm#XXX]

## Summary

One paragraph explanation of the proposal.

## Motivation

Why are we doing this? What use cases does it support? What problems does it solve?

## Detailed Design

Explain the design in sufficient detail that someone familiar with the project can implement it.

## Drawbacks

Why should we *not* do this?

## Alternatives

What other designs have been considered? What is the impact of not doing this?

## Unresolved Questions

What parts of the design are still TBD?
```

### 4.3 Conflict Resolution

**If Core Maintainers Disagree:**
1. Extended discussion period (additional 1-2 weeks)
2. Request community input via GitHub Discussion
3. If still no consensus: Project Lead makes final decision
4. Dissenting opinions documented in RFC

**If Community Objects to Core Maintainer Decision:**
1. Community members can request RFC reconsideration
2. If 10+ community members request (with substantive concerns), Core Maintainers must re-open discussion
3. If still disagreement: Project Lead makes final decision
4. Dissenting community members may fork (open-source nature allows)

**Code of Conduct Violations:**
- Handled by Community Manager (warnings, temporary bans)
- Permanent bans require Core Maintainer approval
- Appeals process: Email core-team@acm.dev with explanation; Core Maintainers review within 1 week

---

## 5. Contribution Guidelines

### 5.1 How to Contribute

**Types of Contributions:**
1. **Code** (features, bug fixes, performance improvements)
2. **Documentation** (guides, tutorials, API docs, translations)
3. **Testing** (write tests, test new features, report bugs)
4. **Design** (UI/UX, branding, graphics)
5. **Community Support** (answer questions on Discord/GitHub, triage issues)
6. **Legal/Compliance** (review EULA, improve Legal NLP, provide legal guidance)
7. **Security** (responsible vulnerability disclosure, security audits)

**Before Contributing:**
1. Read `CONTRIBUTING.md` (comprehensive contribution guide)
2. Check `GOOD_FIRST_ISSUES` label on GitHub for beginner-friendly tasks
3. Introduce yourself in #introductions on Discord (optional but encouraged)
4. Review Code of Conduct

### 5.2 Code Contribution Workflow

```
1. FIND OR CREATE ISSUE
   ├─ Browse open issues: https://github.com/acm-project/acm/issues
   ├─ If no issue exists, create one describing the problem or feature
   └─ Comment on issue to indicate you're working on it (avoid duplicate work)

2. FORK AND BRANCH
   ├─ Fork acm repository to your GitHub account
   ├─ Create feature branch: git checkout -b feature/your-feature-name
   └─ Follow branch naming convention: feature/*, bugfix/*, docs/*

3. DEVELOP
   ├─ Write code following style guide (see Section 5.3)
   ├─ Add tests (unit tests required for new features)
   ├─ Ensure all tests pass: make test
   ├─ Run linters: make lint
   └─ Update documentation if API changed

4. COMMIT
   ├─ Follow conventional commits: feat:, fix:, docs:, chore:, etc.
   ├─ Example: "feat(crs): add LastPass CLI integration"
   └─ Reference issue: "Closes #123"

5. PULL REQUEST
   ├─ Push to your fork
   ├─ Create Pull Request to main branch
   ├─ Fill out PR template (description, testing, screenshots if UI)
   ├─ Link related issue
   └─ Request review from Core Maintainers

6. CODE REVIEW
   ├─ Respond to review comments within 48 hours (or indicate if delayed)
   ├─ Make requested changes
   ├─ Re-request review after updates
   └─ Do not force-push after review started (makes tracking changes difficult)

7. MERGE
   ├─ Requires 2+ Core Maintainer approvals (1 for trivial changes)
   ├─ All CI checks must pass
   ├─ Squash and merge (maintains clean commit history)
   └─ Contributor credited in CHANGELOG and release notes
```

### 5.3 Code Quality Standards

**Go Code Style:**
- Follow official Go conventions: `go fmt`, `gofmt -s`
- Use `golangci-lint` with project configuration
- Maximum function complexity: 15 (cyclomatic complexity)
- Test coverage: > 80% for new code
- Comment exported functions and types (GoDoc)

**TypeScript/React Code Style:**
- Use Prettier for formatting
- ESLint with project configuration
- TypeScript strict mode enabled
- React functional components with hooks (no class components)

**Commit Message Convention:**

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:** feat, fix, docs, style, refactor, test, chore, perf, ci

**Examples:**
```
feat(crs): add support for LastPass CLI integration

- Implement LastPass adapter following abstraction pattern
- Add integration tests for LastPass operations
- Update documentation with LastPass setup instructions

Closes #42
```

```
fix(acvs): correct NLP model parsing of ambiguous ToS clauses

NLP model was incorrectly classifying "automated access with permission"
as prohibited. Updated training data and retrained model.

Fixes #156
```

### 5.4 Testing Requirements

**Before PR Submission:**
- [ ] All existing tests pass: `make test`
- [ ] New features include unit tests
- [ ] Integration tests updated if public API changed
- [ ] Manual testing performed (provide testing notes in PR)
- [ ] No regressions in existing functionality

**Test Coverage Targets:**
- **Core CRS/ACVS logic:** > 90% coverage
- **API handlers:** > 85% coverage
- **CLI commands:** > 80% coverage
- **Overall project:** > 80% coverage

**CI/CD Pipeline:**
- Unit tests run on every PR
- Integration tests run on every PR to `main`
- Security scans (Semgrep, gosec) run on every PR
- Build artifacts generated for successful merges to `main`

### 5.5 Documentation Standards

**Required Documentation for New Features:**
1. **User-facing docs** (in `docs/`)
   - Feature overview and use cases
   - Setup and configuration
   - Examples and tutorials
   - Troubleshooting

2. **API documentation** (GoDoc comments for Go, JSDoc for TypeScript)
3. **Architecture Decision Records (ADRs)** for significant technical decisions
4. **CHANGELOG entry** for user-visible changes

**Documentation Style:**
- Clear, concise, jargon-free language
- Include code examples with expected output
- Use screenshots for UI features
- Provide both CLI and GUI instructions where applicable

---

## 6. Community and Communication

### 6.1 Communication Channels

| Channel | Purpose | Link | Moderation |
|---------|---------|------|------------|
| **GitHub Issues** | Bug reports, feature requests | [github.com/acm-project/acm/issues](https://github.com) | Core Maintainers + Community Manager |
| **GitHub Discussions** | Questions, ideas, general discussion | [github.com/acm-project/acm/discussions](https://github.com) | Community Manager |
| **Discord** | Real-time chat, community support | [discord.gg/acm](https://discord.gg) | Community Manager + Moderators |
| **Mailing List** | Security advisories, announcements | security@acm.dev, announce@acm.dev | Core Maintainers |
| **Blog** | Release notes, technical deep-dives | [acm-project.dev/blog](https://acm-project.dev) | Core Maintainers + Contributors |
| **Twitter/X** | Project updates, engagement | [@acm_project](https://twitter.com) | Community Manager |

### 6.2 Community Events

#### Monthly Community Call

**Schedule:** First Thursday of each month, 5:00 PM UTC  
**Duration:** 60 minutes  
**Format:** Video call (Zoom or Jitsi)  
**Recording:** Published to YouTube (acm-project channel)

**Agenda:**
1. **Welcome and Introductions** (5 min)
   - New contributors introduce themselves
2. **Project Updates** (15 min)
   - Recent releases, roadmap progress
   - Security/legal updates
3. **RFC Discussions** (20 min)
   - Present and discuss active RFCs
   - Community feedback on proposals
4. **Open Forum** (15 min)
   - Community questions and discussions
5. **Contributor Spotlight** (5 min)
   - Recognize outstanding contributor of the month

**Participation:** Open to all; no registration required

---

#### Quarterly Contributor Summit

**Schedule:** Last Saturday of March, June, September, December  
**Duration:** 4 hours (virtual)  
**Format:** Workshops, hackathon, roadmap planning

**Activities:**
- Technical workshops (e.g., "Building a Password Manager Adapter")
- Security deep-dive sessions
- Legal NLP model training workshop
- Roadmap prioritization (community voting)
- Virtual happy hour (last 30 minutes)

---

#### Annual ACM Conference (Future)

**Target:** Year 2+  
**Format:** In-person (if feasible) or hybrid  
**Duration:** 2 days

**Goals:**
- Celebrate project milestones
- Technical talks from contributors and users
- Security and legal workshop tracks
- Contributor awards and recognition

---

### 6.3 Recognition and Incentives

**Contributor Recognition:**

| Recognition | Criteria | Reward |
|-------------|----------|--------|
| **First Contribution Badge** | Merge first PR | GitHub badge, Discord role, featured in newsletter |
| **Contributor of the Month** | Outstanding contribution (nominated by community) | Blog post spotlight, special Discord role, swag (if budget allows) |
| **Core Contributor Badge** | 10+ merged PRs | GitHub badge, listed on website, invited to Core Contributor sync calls |
| **Hall of Fame** | Sustained, exceptional contributions | Permanent recognition on website, named in release notes |

**Swag (if budget allows):**
- Stickers, t-shirts, laptop stickers
- Distributed at conferences or mailed to top contributors
- Funded via donations or sponsorships

**Non-Material Incentives:**
- Professional development: Mentorship from Core Maintainers
- Networking: Access to security/legal experts in community
- Portfolio building: High-quality open-source contributions
- Leadership opportunities: Working Group leads, mentorship roles

---

### 6.4 Code of Conduct

ACM adopts the **Contributor Covenant Code of Conduct v2.1** (https://www.contributor-covenant.org/).

**Key Provisions:**

**Our Pledge:**
- Foster an open, welcoming, inclusive environment
- Treat all contributors with respect, regardless of background

**Expected Behavior:**
- Use welcoming and inclusive language
- Respect differing viewpoints and experiences
- Accept constructive criticism gracefully
- Focus on what's best for the community

**Unacceptable Behavior:**
- Harassment, discrimination, or personal attacks
- Trolling, insulting comments, or deliberate derailment
- Publishing others' private information without permission
- Other conduct reasonably considered inappropriate

**Enforcement:**
1. **Warning:** First minor violation
2. **Temporary Ban (1-7 days):** Repeated or moderate violation
3. **Permanent Ban:** Severe or repeated violations

**Reporting:** Email community@acm.dev or DM Community Manager on Discord  
**Response Time:** Within 48 hours

---

## 7. Security and Vulnerability Disclosure

### 7.1 Responsible Disclosure Policy

**ACM is committed to security and welcomes responsible vulnerability disclosures.**

**How to Report:**
- **Email:** security@acm.dev (PGP key available at https://acm-project.dev/pgp)
- **GitHub Security Advisory:** https://github.com/acm-project/acm/security/advisories
- **Encrypted communication preferred** (PGP or Signal)

**What to Include:**
- Detailed description of vulnerability
- Steps to reproduce (proof-of-concept code if applicable)
- Potential impact and affected versions
- Any proposed fixes or mitigations

**What NOT to Do:**
- Do not publicly disclose vulnerability before fix released
- Do not exploit vulnerability beyond proof-of-concept
- Do not access user data or compromise systems

**Our Commitment:**
- **Acknowledgment:** Within 48 hours of report
- **Initial Assessment:** Within 1 week (P0/P1 severity)
- **Fix Timeline:**
  - P0 (Critical): 24-48 hours
  - P1 (High): 1 week
  - P2 (Medium): 2-4 weeks
- **Credit:** Public acknowledgment in security advisory (if desired by reporter)
- **Coordination:** Work with reporter on coordinated disclosure

**Bug Bounty (Future):**
- If project secures funding, establish formal bug bounty program
- Payouts based on severity (CVSS score)
- Managed via platform (HackerOne or Bugcrowd)

---

### 7.2 Security Audit Program

**Annual Security Audit:**
- Engage third-party security firm for comprehensive audit
- Scope: Core ACM service, CRS/ACVS modules, client-server communication
- Publish audit report (redacted if necessary) on website
- Address findings within defined SLA (P0: 48h, P1: 1 week)

**Community Security Review:**
- Quarterly "Security Sprint" where community focuses on security testing
- Bug bash events with recognition for vulnerability discoveries
- Maintain public list of past vulnerabilities and fixes (CVE database)

---

## 8. Legal and Compliance Governance

### 8.1 Legal Review Process

**For EULA or Legal Documentation Changes:**

1. **Draft**: Legal Review Committee drafts proposed changes
2. **External Review** (for major changes): Engage external legal counsel
3. **Community RFC**: Publish RFC for community feedback (2-week minimum discussion)
4. **Final Review**: Legal Review Committee + Core Maintainers approve
5. **Versioning**: Increment EULA version number; log all changes
6. **User Notification**: Prompt users to re-accept EULA on next run (if material changes)

**For ACVS Legal NLP Model Updates:**

1. **Quarterly Accuracy Review**: Legal Review Committee reviews 100 sample ToS analyses
2. **Error Identification**: Document false positives/negatives
3. **Model Retraining**: Legal NLP WG retrains model with corrected data
4. **Validation**: Test on holdout set; require > 85% F1 score
5. **Release**: Tag new model version; document changes in release notes

### 8.2 Incident Response for Legal Threats

**If Project or User Receives Legal Threat:**

1. **Immediate Notification**: Contact legal@acm.dev (monitored by Legal Review Committee + Project Lead)
2. **Assessment**: Legal Review Committee assesses threat severity
3. **External Counsel** (if serious): Engage attorney (if project has legal insurance or pro bono counsel)
4. **Response Strategy**:
   - User threat: Point to EULA indemnification; provide evidence chain export
   - Project threat: Assert EULA, open-source disclaimers, ACVS good-faith effort
5. **Communication**: Internal only until resolved; public post-mortem if appropriate
6. **Update Risk Register**: Add learnings to risk assessment document

---

## 9. Financial and Sustainability Model

### 9.1 Current Funding Model

**Phase I (MVP):** Volunteer-driven, no external funding

**Costs:**
- Domain registration ($15/year)
- Infrastructure (CI/CD, hosting for website): ~$50/month (covered by contributors or free tiers)
- Legal consultation (optional): Pro bono or deferred

**Total Estimated Annual Cost:** < $1,000

---

### 9.2 Future Funding Strategies

**If Project Scales Beyond Volunteer Capacity:**

1. **Donations**
   - GitHub Sponsors
   - Open Collective
   - Patreon
   - Cryptocurrency donations (Bitcoin, Ethereum)

2. **Grants**
   - Open Source Security Foundation (OpenSSF)
   - Mozilla Open Source Support (MOSS)
   - Sovereign Tech Fund (EU)

3. **Corporate Sponsorships**
   - Bronze/Silver/Gold tiers
   - Logo placement on website and README
   - Priority support (not exclusive features)

4. **Consulting/Training** (optional, if demand exists)
   - Paid workshops on credential security
   - Custom ACVS training for enterprise deployments
   - Revenue supports core project development

**Principles:**
- **No ads or tracking** in the software (ever)
- **No feature paywalls** — all features remain open-source
- **Transparent finances** — publish budget and spending reports
- **Community governance** — major spending decisions require community approval

---

### 9.3 Budget Allocation (If Funded)

| Category | % of Budget | Use Cases |
|----------|-------------|-----------|
| **Infrastructure** | 20% | Hosting, CI/CD, domain, email |
| **Legal** | 30% | External counsel, E&O insurance (if available) |
| **Security** | 30% | Annual audits, bug bounty program, security tools |
| **Community** | 10% | Swag, event sponsorships, contributor travel |
| **Contingency/Reserves** | 10% | Emergency legal or security needs |

---

## 10. Project Roadmap

### 10.1 Development Phases

#### Phase I: MVP — Credential Remediation Service (CRS)

**Timeline:** Months 1-4 (Target: Q1 2026)  
**Status:** 🔴 Not Started

**Goals:**
- Establish zero-knowledge foundation with password manager CLI integration
- Implement local breach detection via built-in password manager reports
- Enable basic rotation workflow with HIM (Human-in-the-Middle)
- Release as alpha/beta for early adopters

**Key Deliverables:**

| Deliverable | Owner | Status | Target Date |
|-------------|-------|--------|-------------|
| **Core ACM Service (Go)** | Technical Lead | Not Started | Month 2 |
| CRS module with 1Password integration | Technical Lead | Not Started | Month 2 |
| CRS module with Bitwarden integration | Technical Lead | Not Started | Month 3 |
| OpenTUI client (basic functionality) | Technical Lead | Not Started | Month 3 |
| mTLS authentication with certificates | Security Lead | Not Started | Month 2 |
| Local audit logging (SQLite) | Technical Lead | Not Started | Month 2 |
| Basic HIM workflow (user manually completes rotation) | Technical Lead | Not Started | Month 3 |
| **Documentation** | Documentation WG | Not Started | Month 4 |
| Installation and setup guide | Documentation WG | Not Started | Month 4 |
| Architecture documentation (TAD) | Technical Lead | Not Started | Month 4 |
| Security threat model | Security Lead | Not Started | Month 4 |
| **Legal Framework** | Legal Review Committee | Not Started | Month 4 |
| EULA v1.0 (reviewed by counsel) | Legal Advisor | Not Started | Month 3 |
| EULA acceptance flow implementation | Technical Lead | Not Started | Month 4 |
| **Testing and Quality** | QA Team (volunteer) | Not Started | Month 4 |
| Unit test coverage > 80% | All Contributors | Not Started | Month 4 |
| Integration tests for CLI interactions | Technical Lead | Not Started | Month 4 |
| Security review (community) | Security Lead | Not Started | Month 4 |
| **Release** | Project Lead | Not Started | Month 4 |
| Beta release (GitHub Releases) | Project Lead | Not Started | Month 4 |
| Announcement blog post | Community Manager | Not Started | Month 4 |

**Success Criteria:**
- ✅ 50+ beta users successfully install and use ACM
- ✅ Zero critical security vulnerabilities discovered post-release
- ✅ 80%+ unit test coverage
- ✅ EULA legally reviewed and accepted by users
- ✅ Positive community feedback on Discord/GitHub

---

#### Phase II: ACVS — Automated Compliance Validation Service

**Timeline:** Months 5-8 (Target: Q2 2026)  
**Status:** 🔴 Not Started

**Goals:**
- Introduce automated ToS compliance validation
- Implement Legal NLP engine for ToS parsing
- Enable API-based rotation for services with documented APIs
- Generate evidence chains for compliance proof

**Key Deliverables:**

| Deliverable | Owner | Status | Target Date |
|-------------|-------|--------|-------------|
| **ACVS Module** | Legal NLP WG + Technical Lead | Not Started | Month 7 |
| Legal NLP engine (spaCy or Transformers) | Legal NLP WG | Not Started | Month 6 |
| ToS parsing and CRC generation | Legal NLP WG | Not Started | Month 6 |
| Compliance validation logic | Technical Lead | Not Started | Month 7 |
| Evidence chain system (cryptographically signed) | Technical Lead | Not Started | Month 7 |
| **API-Based Rotation** | Password Manager Integration WG | Not Started | Month 7 |
| GitHub API integration (example) | PM Integration WG | Not Started | Month 6 |
| AWS IAM credential rotation (example) | PM Integration WG | Not Started | Month 7 |
| Google Account API (if available) | PM Integration WG | Not Started | Month 7 |
| **Enhanced HIM Workflow** | Technical Lead | Not Started | Month 8 |
| TOTP/SMS MFA support | Technical Lead | Not Started | Month 8 |
| CAPTCHA handling (secure browser view) | Technical Lead | Not Started | Month 8 |
| **User Interface** | UX/Design WG | Not Started | Month 8 |
| ACVS opt-in flow in OpenTUI | UX/Design WG | Not Started | Month 7 |
| Tauri GUI (basic functionality) | UX/Design WG | Not Started | Month 8 |
| Compliance dashboard (visualize CRC status) | UX/Design WG | Not Started | Month 8 |
| **Legal Updates** | Legal Review Committee | Not Started | Month 8 |
| EULA v2.0 with ACVS provisions | Legal Advisor | Not Started | Month 7 |
| ACVS re-acceptance workflow | Technical Lead | Not Started | Month 8 |
| **Documentation** | Documentation WG | Not Started | Month 8 |
| ACVS user guide | Documentation WG | Not Started | Month 8 |
| Legal NLP model documentation | Legal NLP WG | Not Started | Month 8 |
| Evidence chain export guide | Documentation WG | Not Started | Month 8 |
| **Release** | Project Lead | Not Started | Month 8 |
| Stable v1.0 release | Project Lead | Not Started | Month 8 |
| Security audit (third-party) | Security Lead | Not Started | Month 8 |

**Success Criteria:**
- ✅ Legal NLP model achieves > 85% F1 score on test set
- ✅ API-based rotation works for 5+ major services
- ✅ Evidence chain cryptographically verifiable
- ✅ Third-party security audit with no critical findings
- ✅ 500+ users adopt ACVS (opt-in)

---

#### Phase III: Advanced Automation & Ecosystem Growth

**Timeline:** Months 9-12 (Target: Q3-Q4 2026)  
**Status:** 🔴 Not Started

**Goals:**
- Expand automation capabilities with controlled UI scripting
- Improve Legal NLP model accuracy and coverage
- Grow community and ecosystem (plugins, integrations)
- Stabilize for production use

**Key Deliverables:**

| Area | Deliverables |
|------|-------------|
| **UI Scripting** | • Browser automation framework (Playwright)<br>• ACVS-validated UI scripting for low-risk sites<br>• Enhanced HIM for hardware security keys (FIDO2) |
| **Performance** | • Optimize NLP inference (GPU acceleration)<br>• Reduce memory footprint<br>• Improve startup time |
| **Integrations** | • Additional password managers (LastPass, KeePassXC)<br>• SSO providers (Okta, Auth0) for enterprise<br>• Cloud IAM rotation (AWS, Azure, GCP) |
| **Community** | • 100+ contributors<br>• 10+ active Working Groups<br>• Quarterly Contributor Summit established |
| **Documentation** | • Multi-language docs (Spanish, French, German)<br>• Video tutorials<br>• Interactive onboarding |

**Success Criteria:**
- ✅ 5,000+ active users
- ✅ 100+ contributors
- ✅ Stable v2.0 release with no critical bugs
- ✅ Featured in major security publications (e.g., Ars Technica, The Hacker News)

---

#### Phase IV: Enterprise & Ecosystem Maturity

**Timeline:** Year 2+  
**Status:** 🔴 Future

**Goals:**
- Enterprise deployment support
- Hardware security module (HSM) integration
- Federated Legal NLP model sharing
- Sustainable funding model established

**Key Features:**
- Centralized policy management for organizations
- TPM/Secure Enclave integration for certificate storage
- Browser extension for real-time breach alerts
- Mobile app (iOS/Android)
- Annual ACM Conference

---

### 10.2 Roadmap Prioritization

**How We Prioritize:**

1. **Security > Everything Else**: Security vulnerabilities always top priority
2. **User-Facing Impact**: Features that benefit most users prioritized
3. **Community Input**: RFC voting and GitHub Discussions inform priorities
4. **Resource Availability**: Dependent on contributor time and expertise
5. **Risk Mitigation**: High-risk items from Risk Assessment addressed early

**Quarterly Roadmap Review:**
- Community votes on priorities for next quarter
- Core Maintainers finalize roadmap based on feasibility
- Published on GitHub Projects board (public)

---

## 11. Metrics and Success Indicators

### 11.1 Project Health Metrics

| Metric | Target (Year 1) | Target (Year 2) | Measurement |
|--------|----------------|-----------------|-------------|
| **Active Contributors** | 50+ | 150+ | GitHub Insights |
| **Total Contributors** | 100+ | 300+ | GitHub Insights |
| **Pull Requests (Monthly)** | 20+ | 50+ | GitHub API |
| **Active Users** | 1,000+ | 10,000+ | Opt-in telemetry |
| **GitHub Stars** | 2,000+ | 10,000+ | GitHub |
| **Discord Members** | 500+ | 2,000+ | Discord |
| **Security Vulnerabilities (Open)** | 0 critical, < 3 high | 0 critical, < 2 high | GitHub Security |
| **Test Coverage** | > 80% | > 85% | Codecov |
| **Community Sentiment** | > 80% positive | > 85% positive | Surveys, Discord sentiment |

### 11.2 Community Health Indicators

**Leading Indicators of Healthy Community:**
- Active participation in RFCs (10+ comments per RFC)
- Regular new contributor onboarding (5+ per month)
- Low Code of Conduct violations (< 1 per month)
- High contributor retention (50%+ return contributors)
- Diverse contributor base (geographic, skill levels, backgrounds)

**Red Flags:**
- Declining PR submissions
- Increasing issue backlog (> 100 open issues)
- Toxic behavior in community spaces
- Maintainer burnout (slow response times)

**Mitigation for Red Flags:**
- Recruit new maintainers
- Organize issue triage sprint
- Enforce Code of Conduct strictly
- Provide maintainer support and recognition

---

## 12. Transition and Succession Planning

### 12.1 Maintainer Succession

**If Project Lead Steps Down:**
1. Project Lead nominates successor OR Core Maintainers vote
2. Community feedback period (2 weeks)
3. Formal handoff with documentation transfer
4. Announcement in blog post and community call

**If Core Maintainer Inactive (> 3 months):**
1. Community Manager contacts maintainer to check status
2. If no response, consider temporary leave or removal
3. Recruit replacement from active contributors
4. Document knowledge transfer

### 12.2 Project Archival (If Necessary)

**If Project Becomes Unmaintainable:**
- Archive GitHub repository (read-only)
- Publish final security advisory recommending alternatives
- Transfer domain and assets to successor project (if exists)
- Document reasons for archival transparently

**Commitment:** Project Lead commits to 2+ years of active maintenance before considering archival.

---

## 13. Conclusion

### 13.1 Call to Action

**ACM is at the beginning of an exciting journey.**

We invite:
- **Developers** to contribute code, tests, and documentation
- **Security professionals** to audit, test, and provide feedback
- **Legal experts** to join the Legal Review Committee
- **Designers** to improve UX and create visual assets
- **Users** to adopt, provide feedback, and spread the word

**Get Involved:**
- Star the repository: https://github.com/acm-project/acm
- Join Discord: https://discord.gg/acm
- Read CONTRIBUTING.md and find a "good first issue"
- Follow @acm_project on Twitter/X

### 13.2 Acknowledgments

This governance framework and roadmap are inspired by successful open-source projects:

- **Kubernetes**: Steering Committee and SIG (Special Interest Group) model
- **Rust**: RFC process and community-driven decision-making
- **1Password**: Password manager security best practices
- **Bitwarden**: Open-source password management architecture
- **Local-First Software**: Zero-knowledge and privacy-first principles

**Special Thanks** (to be filled as project grows):
- Initial contributors
- Legal advisors
- Security researchers
- Early adopters

---

## Document History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2025-11-13 | Initial Draft | Created governance framework and roadmap |
| 1.0 | 2025-11-13 | Claude (AI Assistant) | Complete governance structure with RFC process, contribution guidelines, and 4-phase roadmap |

---

**Document Status:** Draft — Requires Community Input  
**Next Review Date:** [Upon project initialization]  
**Distribution:** Public (Open-Source Project Documentation)

---

**Let's build something great together. Welcome to ACM.** 🚀
