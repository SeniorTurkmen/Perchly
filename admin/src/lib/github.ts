// Server-side only — GITHUB_TOKEN never reaches the browser, same
// reasoning as BACKEND_URL (see lib/backend.ts). Configure both in
// admin/.env.local: GITHUB_TOKEN (a personal access token with at
// least read access to the repo — "repo" scope for a private repo,
// "public_repo" is enough for a public one) and GITHUB_REPO
// ("owner/name", e.g. "SeniorTurkmen/Perchly").
const GITHUB_TOKEN = process.env.GITHUB_TOKEN ?? "";
const GITHUB_REPO = process.env.GITHUB_REPO ?? "";

export const isGitHubConfigured = Boolean(GITHUB_REPO);

export class GitHubApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function githubFetch<T>(path: string): Promise<T> {
  if (!isGitHubConfigured) {
    throw new GitHubApiError("GitHub yapılandırılmamış", 0);
  }

  const headers: Record<string, string> = {
    Accept: "application/vnd.github+json",
    "X-GitHub-Api-Version": "2022-11-28",
  };
  if (GITHUB_TOKEN) headers.Authorization = `Bearer ${GITHUB_TOKEN}`;

  const res = await fetch(`https://api.github.com/repos/${GITHUB_REPO}${path}`, {
    headers,
    cache: "no-store",
  });
  if (!res.ok) {
    if (res.status === 404) {
      throw new GitHubApiError(
        "Repo bulunamadı — özel bir repoysa GITHUB_TOKEN gerekir.",
        404,
      );
    }
    if (res.status === 401 || res.status === 403) {
      throw new GitHubApiError("GitHub kimlik doğrulaması başarısız (token geçersiz veya izin yetersiz).", res.status);
    }
    throw new GitHubApiError(`GitHub API hatası (${res.status})`, res.status);
  }
  return res.json();
}

export type GitHubUser = {
  login: string;
  avatar_url: string;
  html_url: string;
};

export type GitHubLabel = {
  name: string;
  color: string;
};

export type GitHubIssue = {
  number: number;
  title: string;
  html_url: string;
  state: "open" | "closed";
  labels: GitHubLabel[];
  user: GitHubUser | null;
  comments: number;
  created_at: string;
  updated_at: string;
  // Present (even if null) only on pull requests — the /issues endpoint
  // returns both issues and PRs, this is how the two are told apart.
  pull_request?: unknown;
};

export type GitHubPullRequest = {
  number: number;
  title: string;
  html_url: string;
  state: "open" | "closed";
  draft: boolean;
  user: GitHubUser | null;
  merged_at: string | null;
  created_at: string;
  updated_at: string;
};

export async function listOpenIssues(): Promise<GitHubIssue[]> {
  const issues = await githubFetch<GitHubIssue[]>(
    "/issues?state=open&per_page=20&sort=updated",
  );
  return issues.filter((issue) => !issue.pull_request);
}

export async function listRecentPullRequests(): Promise<GitHubPullRequest[]> {
  return githubFetch<GitHubPullRequest[]>("/pulls?state=all&per_page=10&sort=updated");
}

export type GitHubRepoSummary = {
  full_name: string;
  html_url: string;
  open_issues_count: number;
  default_branch: string;
  pushed_at: string;
};

export async function getRepoSummary(): Promise<GitHubRepoSummary> {
  return githubFetch<GitHubRepoSummary>("");
}
