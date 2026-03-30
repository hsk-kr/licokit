# When to Invoke Agents vs Follow Skills

## Decision Tree

```
Is this a coding task?
├── YES: Does it involve writing/editing implementation code?
│   ├── YES: Is it a trivial single-line edit (typo, import)?
│   │   ├── YES → Do it directly (no agent)
│   │   └── NO → Invoke the matching agent via Agent tool
│   └── NO: Is it research/analysis?
│       ├── YES → Use codebase-analyzer, codebase-locator, or codebase-pattern-finder
│       └── NO → Handle directly
└── NO: Is it documentation, planning, or review?
    ├── YES → Use the matching role agent (documentation-writer, etc.)
    └── NO → Handle directly
```

## Key Distinctions

### Skills = Knowledge to Follow
- Skills are **instructions** that tell you HOW to do something
- You follow skills yourself — they don't spawn separate agents
- Example: "code-review" skill tells you the review checklist

### Agents = Workers to Invoke
- Agents are **autonomous workers** you delegate tasks to
- You invoke them via the Agent tool with a specific prompt
- Example: "typescript-pro" agent writes the actual TS code

### When Both Apply
- If a skill and agent cover the same domain, invoke the agent AND tell it to follow the skill
- Example: For React work, invoke `react-specialist` agent

## Agent Quick Reference

| Situation | Agent(s) |
|---|---|
| Writing Go code | `golang-pro` |
| Writing React component | `react-specialist` |
| Writing TypeScript | `typescript-pro` |
| Writing Python | `python-pro` |
| Writing JavaScript | `javascript-pro` |
| Writing Swift | `swift-expert` |
| Building backend API | `backend-developer` |
| Building full-stack feature | `fullstack-developer` |
| Building frontend | `frontend-developer` |
| Building mobile app | `mobile-developer` |
| Designing an API | `api-designer` |
| Designing UI | `ui-designer` |
| Need to find a file | `codebase-locator` |
| Need to understand code | `codebase-analyzer` |
| Need to find patterns | `codebase-pattern-finder` |
| Neovim/Lua work | `neovim-lua` |
| Web research | `web-search-researcher` |
| Writing tests | `test-engineer` |
| Security review | `security-auditor` |
| Debugging a bug | `debugging-detective` |
| Safe refactoring | `refactoring-surgeon` |
| Database design/queries | `database-architect` |
| CI/CD/Docker/K8s | `devops-engineer` |
| Dependency management | `dependency-manager` |
| Complex git operations | `git-strategist` |
| Performance optimization | `performance-optimizer` |
| Writing documentation | `documentation-writer` |
| Fixing a typo | Do it directly |
| Adding an import | Do it directly |

## Team Patterns — Cross-Checking Agents

The real power is using agents in **teams** where they cross-check each other. Use `TeamCreate` or parallel `Agent` calls for these patterns.

### Pattern 1: Write → Test → Secure (Every Feature)
```
1. Developer agent writes the code (typescript-pro, python-pro, etc.)
2. IN PARALLEL:
   - test-engineer: writes tests for the new code
   - security-auditor: audits the new code for vulnerabilities
3. Fix any issues found
```

### Pattern 2: Debug → Fix → Verify (Every Bug)
```
1. IN PARALLEL:
   - debugging-detective: traces the root cause
   - codebase-analyzer: maps the component architecture
2. Developer agent writes the fix
3. IN PARALLEL:
   - test-engineer: writes regression test
   - security-auditor: verifies fix doesn't introduce vulns
```

### Pattern 3: Refactor → Verify → Review (Every Refactoring)
```
1. test-engineer: writes characterization tests (if missing)
2. refactoring-surgeon: performs the refactoring in small steps
3. IN PARALLEL:
   - test-engineer: verifies all tests pass
   - security-auditor: checks for security regressions
   - performance-optimizer: checks for perf regressions
```

### Pattern 4: Design → Build → Ship (New Feature)
```
1. IN PARALLEL:
   - codebase-analyzer: understand existing architecture
   - codebase-pattern-finder: find similar features to model after
2. api-designer or database-architect: design the data/API layer
3. Developer agents IN PARALLEL: build frontend + backend
4. IN PARALLEL:
   - test-engineer: write tests
   - security-auditor: audit
   - documentation-writer: document the new feature
```

### Pattern 5: Dependency Upgrade (Maintenance)
```
1. dependency-manager: audit and identify upgrades
2. Developer agent: apply the upgrade
3. IN PARALLEL:
   - test-engineer: run full test suite
   - security-auditor: verify no new vulns introduced
```

### Pattern 6: PR Review (Quality Gate)
```
IN PARALLEL (all at once):
- security-auditor: check for vulnerabilities
- test-engineer: review test coverage and quality
- performance-optimizer: check for perf regressions
- code-review skill: general code quality
```

## When NOT to Use Agents

- Trivial edits (typos, imports, single-line fixes) — do it directly
- Reading a single file — use Read tool
- Simple grep — use Grep tool
- Tasks that take < 30 seconds — do it directly
