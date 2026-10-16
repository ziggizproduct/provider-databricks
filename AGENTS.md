# Agent guide

<!-- standards:begin (managed by ziggizproduct/arch scripts/sync-standards.sh; edit there) -->

## Formatting and linting

This repo follows the workspace standard in
[ziggizproduct/arch](https://github.com/ziggizproduct/arch/blob/main/docs/standards/engineering/formatting-and-linting.md).
Two gates enforce it: the pre-commit git hook on every commit, and the
required `lint / pre-commit` check on every pull request (merges to the
default branch are blocked until it passes).

- At the start of every session, run `scripts/ensure-pre-commit.sh` in each
  repo you'll commit to; it installs pre-commit and the git hook. Never bypass the hook with
  `git commit --no-verify`.

- Before every commit run `pre-commit run --all-files` (or `--files <paths>`
  for just what you changed). Hooks fix what they can in place; re-run until
  it passes, then stage the fixes with your change.
- Never hand-format. Never edit the shared config files (`.pre-commit-config.yaml`,
  `biome.json`, `ruff.toml`, `.golangci.yml`, `rustfmt.toml`, `clippy.toml`,
  `.tflint.hcl`, `.yamllint.yaml`, `.markdownlint-cli2.yaml`, `.prettierrc.yaml`,
  `.editorconfig`, `scripts/ensure-pre-commit.sh`, `.github/workflows/lint.yml`,
  `.github/workflows/commits.yml`);
  change them in the arch repo.
- Fix the code rather than suppress a rule. If a suppression is unavoidable,
  scope it to one line, name the rule and give the reason (forms per tool are
  in the standard's "Suppressing a finding" table).
- A hook that reports `Executable ... not found` needs a tool from the
  standard's "Tools you install" table; install it at the listed version.

## Commit messages

Commits and pull request titles follow
[Conventional Commits](https://github.com/ziggizproduct/arch/blob/main/docs/standards/engineering/commits.md):
`<type>[(scope)][!]: <description>`, types `feat fix docs chore refactor test
ci build perf revert style`. The PR title becomes the commit on `main`; the
`commit-msg` hook and the `commits / conventional` check enforce it.

## Agent roles and senior review

This repo follows the
[agent roles standard](https://github.com/ziggizproduct/arch/blob/main/docs/standards/way-of-working/roles.md).

- Three senior agents (`architect`, `program-manager`, `dev-lead`), the
  workspace-wide specialists and the `senior-review` skill ship in arch's
  `ziggiz-agents` plugin, which this repo's `.claude/settings.json` and the
  workspace directory's enable. Invoke them as `ziggiz-agents:<name>`. Start
  sessions in the workspace directory, the parent of the repo checkouts
  ([setting up a workspace](https://github.com/ziggizproduct/arch/blob/main/docs/standards/way-of-working/worktrees-and-prs.md#setting-up-a-workspace)).
- Follow [work, decisions and docs](https://github.com/ziggizproduct/arch/blob/main/docs/standards/way-of-working/work-and-decisions.md).
  Learn how things are from the architecture pages, specs, standards and
  the [stack](https://github.com/ziggizproduct/arch/blob/main/docs/standards/stack.md);
  ADRs say why. Work is epics and features, never an ADR. An ADR is for a
  genuinely contested choice between real alternatives, not for anything
  simply directed or settled as standard practice.
  Use tools and libraries from the stack; add a new one by PR; it's Adopt when that PR merges. A
  question that blocks work becomes a `question` issue with a marker in the
  page its answer will change; put up to four to the product owner at once.
- Track work in GitHub with the `ziggiz-agents:project-tracking` skill:
  start from the highest-scored Ready item, set an issue's type, parent,
  blocked-by links and Effort, Risk reduction and Confidence, move it to In
  Progress with its first pull request, and link every pull request to it.
- Run `senior-review` on every ADR before its pull request merges, on
  every multi-PR plan before work starts, and on every pull request while it
  is a draft. Include this repo's specialists from `.claude/agents/`.
- Before planning any sustained tooling (a custom skill, agent, script,
  hook, workflow, check, controller or rule the team would keep running),
  answer in the plan: is the problem ours or general; if general, which
  tool on the stack, vendor service or open-source tool solves the class
  (write our rule in its language: Rego, `tofu test`, Chainsaw); does that
  tool already check itself (`tofu plan` drift, Crossplane `Synced`/`Ready`)
  so no checker is needed? Custom sustained tooling for a general problem,
  a checker over a self-checking tool, or a tool new to the stack (not
  listed in the stack, and not a package inside a listed ecosystem) goes to
  the owner as one question per plan, at plan time; the team settles the
  rest, including custom tooling for a problem only Ziggiz has.
- To find the tool, list the expert-maintained tools for the problem's domain: agent
  tooling (user skills, installed plugins, marketplace catalogs, this repo's
  `.claude/`, the `ziggiz-agents` plugin, built-in features), the vendor's
  native services (for AWS checks: AWS Config, Security Hub, IAM Access
  Analyzer) and established open-source tools with their built-in rules, on
  the stack or not. Read the ones that cover the task and try the closest. Build something custom only with a written case
  in its design and pull request: what was tried, its output, and why it
  falls short. Reviewers block a custom build with no written case.
- Identity flows down
  ([standard](https://github.com/ziggizproduct/arch/blob/main/docs/standards/engineering/identity-flows-down.md)):
  root (`ziggizproduct/abzyx.dev`) creates every repo's CI identity (AWS role,
  Databricks principal, Cloudflare token). Never create an OIDC-trusted role,
  a federation policy or a token for another repo outside root; to get more
  access, open a PR against your identity file in root.
- Public modules before custom
  ([standard](https://github.com/ziggizproduct/arch/blob/main/docs/standards/engineering/public-modules-first.md)):
  use a pinned `terraform-aws-modules/*` module unless there is a written
  reason; a new `modules/<name>/` README needs a "Why no public module"
  section.
- Every decision names the machine check that enforces it (test, CI check,
  ruleset, policy rule, monitor) from a shared `ci` workflow, or links an
  `enforcement-gap` issue. Never a one-off check for one concern.
- Extend a tool on the stack through its own extension points (config,
  plugins, schemas, documented `--json` output) after trying its commands at
  the pinned version. Never build a competing version of what it does or what
  its maintainers are building; check its releases and open upstream issues
  and pull requests first.
- The review is advisory until the platform is feature complete, except
  that sustained tooling with no plan-time answers, written case or needed
  owner decision blocks the plan, the merge and marking ready. Apply the
  verified steps, do the in-scope briefs, and list the rest as issue
  candidates. Reviewers settle technical questions as findings, by the ADRs,
  lanes, stack and common practice. Only product scope, risk acceptance and
  account-holder questions go to a human; don't answer those yourself.
- Ask the product owner those questions and your own open product and risk
  decisions (never whether to commit, push, mark ready or merge) through the
  structured question prompt (`AskUserQuestion` in Claude Code) with the
  recommended option first. Never ask in prose. With no structured prompt
  (headless, CI, another tool), list the questions in that shape in the PR
  comment or hand-back and stop.
- Keep this repo's checkout on `main`. Do each piece of work in its own
  worktree under the workspace's `.worktrees/<repo>/<branch>`
  (`git worktree add -b <branch> ../.worktrees/<repo>/<branch> origin/main`
  from the repo checkout), on a new branch from `origin/main`, and work in it
  by absolute path from the session in the workspace directory.
  The `dev-lead` subagent reviews and merges every pull request (every repo
  and kind, infra, ADRs and agent rules included) and approves deployments;
  the working agent doesn't, and never asks a person to. Every PR body
  carries one `Intent:` line linking or quoting what the product owner
  asked for; the dev lead merges once the review is posted and the change
  stays within it. Only a question that intent doesn't answer goes to the
  product owner
  ([who merges what](https://github.com/ziggizproduct/arch/blob/main/docs/standards/way-of-working/work-and-decisions.md#who-merges-what)).
  After any PR you worked on merges, without being asked: check CI on `main`
  at the merge commit and fix or report a failure,
  `git fetch --prune && git pull --ff-only` on `main`, delete the worktree
  (`git worktree remove`), confirm the linked issue
  closed and close it if it didn't, and report in one message.
- A pull request you open is a single commit. To update it, amend that commit
  (or squash into it) and force-push the PR branch with
  `git push --force-with-lease`. Never force-push `main`.
- Open every pull request as a draft. Run `senior-review` on the draft,
  apply its fixes, and wait for the checks. Mark it ready only when review
  and development are done: questions answered, checks passing, nothing
  uncommitted or unpushed (`git status --porcelain` empty, local `HEAD` equal
  to the PR head). Ready means the agents are done. Automated standards-sync
  PRs open ready and merge themselves once their required checks pass; only
  that workflow turns on auto-merge, agents never run `gh pr merge --auto`.
- Don't mark a PR ready while you know of a problem in it. Once it's ready,
  don't edit its worktree; local changes don't count. To change it,
  `gh pr ready --undo`, amend, push, wait for checks, mark it ready again.
  Never hand a draft to anyone to merge.
- A PR merges only when review comments are resolved and checks pass.
- Blocked is not a stop. Write the plan that unblocks it (steps, owner per
  step, human actions, time box, done-when), run `senior-review` on it as a
  plan, post it on the blocking issue, do your steps, and put the human
  actions to the human.
- A permission or classifier denial, or a step only a person can take, ends
  the turn with that ask: what to do and what it unblocks. Finish other
  steps first only if they're trivial. Never mention it in passing and
  carry on. Fold a person's edit to an in-flight PR's files
  into that PR's commit.
- End every status report with numbered **Next actions**, each with an owner
  and the default you'll take if nobody answers; at the end of a thread of
  work that default is to wait. Questions go alongside it, never instead
  of it.
- Ask before creating a repo, cloud identity, Argo CD application or other
  new infrastructure resource; before filing an issue (list candidates, file
  the ones picked; a blocking question is filed with arch's Question form
  without asking); and before starting a concept, ADR or design nobody
  asked for (start it as a conversation). When a thread of work ends, stop
  and ask what's next; review follow-ups are candidates, not new work.
- Every claim about state (PR ready or merged, repo exists, issue open, who
  did what) and every recommendation cites evidence gathered in the same
  turn: command output, file and line, or PR or issue field. Check live
  state before asking anyone to delete or decide. Agents use the
  product owner's GitHub token, so a PR's author or `mergedBy` doesn't show
  whether an agent or a person acted; don't use them to say who did.
- A claim about what a tool, service or library does cites the vendor's
  current docs (page, and version or updated date) or a run of the pinned
  version. Training is the last resort; label any claim resting on it as
  unchecked. When evaluating a tool, check its latest release first.
- A directive the product owner gave is settled: carry it out. Don't offer
  it back as a choice, re-argue it, or ask whether to begin. If it looks
  impossible, show the attempt and its output; if two conflict, show both
  and ask which holds.
- Every issue you file has a goal, owner, next action, done-when and
  blocked-by.
- After a PR changing the `ziggiz-agents` plugin merges, start a new
  session; sessions keep the definitions they started with. The `ziggiz`
  marketplace updates itself at session start.
- This repo's `.claude/agents/` holds only its own specialists. When a senior
  asks for one, look for an existing plugin agent first and add one only with
  the written case; workspace-wide specialists go in arch's plugin.

<!-- standards:end -->
