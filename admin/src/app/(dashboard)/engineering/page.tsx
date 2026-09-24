import { LocalDateTime } from "@/components/local-date-time";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { checkBackendHealth } from "@/lib/backend";
import {
  GitHubApiError,
  type GitHubIssue,
  type GitHubPullRequest,
  type GitHubRepoSummary,
  getRepoSummary,
  isGitHubConfigured,
  listOpenIssues,
  listRecentPullRequests,
} from "@/lib/github";

async function BackendHealthCard() {
  let health: { status: string; database: string } | null = null;
  let error: string | null = null;
  try {
    health = await checkBackendHealth();
  } catch {
    error = "Backend'e ulaşılamadı.";
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Backend durumu</CardTitle>
      </CardHeader>
      <CardContent className="flex items-center gap-2 text-sm">
        {health ? (
          <>
            <Badge>Ayakta</Badge>
            <span className="text-muted-foreground">database: {health.database}</span>
          </>
        ) : (
          <>
            <Badge variant="destructive">Erişilemiyor</Badge>
            <span className="text-muted-foreground">{error}</span>
          </>
        )}
      </CardContent>
    </Card>
  );
}

export default async function EngineeringPage() {
  if (!isGitHubConfigured) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold">Mühendislik</h1>
          <p className="text-muted-foreground">GitHub issue/PR&apos;ları ve backend durumu.</p>
        </div>
        <BackendHealthCard />
        <Card>
          <CardHeader>
            <CardTitle>GitHub bağlı değil</CardTitle>
            <CardDescription>
              admin/.env.local dosyasına <code>GITHUB_TOKEN</code> (repo&apos;ya en az okuma
              erişimi olan bir personal access token) ve <code>GITHUB_REPO</code> (örn.{" "}
              <code>SeniorTurkmen/Perchly</code>) ekleyip admin panelini yeniden başlat.
            </CardDescription>
          </CardHeader>
        </Card>
      </div>
    );
  }

  let issues: GitHubIssue[] = [];
  let pullRequests: GitHubPullRequest[] = [];
  let repo: GitHubRepoSummary | null = null;
  let githubError: string | null = null;

  try {
    [repo, issues, pullRequests] = await Promise.all([
      getRepoSummary(),
      listOpenIssues(),
      listRecentPullRequests(),
    ]);
  } catch (err) {
    githubError = err instanceof GitHubApiError ? err.message : "GitHub'dan veri alınamadı.";
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Mühendislik</h1>
        {repo && (
          <p className="text-muted-foreground">
            <a
              href={repo.html_url}
              target="_blank"
              rel="noreferrer"
              className="hover:underline"
            >
              {repo.full_name}
            </a>{" "}
            — {repo.open_issues_count} açık issue/PR · varsayılan dal: {repo.default_branch}
          </p>
        )}
      </div>

      <BackendHealthCard />

      {githubError && (
        <Card>
          <CardHeader>
            <CardTitle>GitHub&apos;a ulaşılamadı</CardTitle>
            <CardDescription>{githubError}</CardDescription>
          </CardHeader>
        </Card>
      )}

      <div className="grid gap-4 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Açık issue&apos;lar</CardTitle>
            <CardDescription>{issues.length} açık issue.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {issues.length === 0 && (
              <p className="text-sm text-muted-foreground">Açık issue yok.</p>
            )}
            {issues.map((issue) => (
              <a
                key={issue.number}
                href={issue.html_url}
                target="_blank"
                rel="noreferrer"
                className="block rounded-md border p-3 text-sm hover:bg-muted/40"
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="font-medium">
                    #{issue.number} {issue.title}
                  </span>
                  <span className="shrink-0 text-xs text-muted-foreground">
                    <LocalDateTime value={issue.updated_at} />
                  </span>
                </div>
                {issue.labels.length > 0 && (
                  <div className="mt-2 flex flex-wrap gap-1">
                    {issue.labels.map((label) => (
                      <Badge
                        key={label.name}
                        variant="outline"
                        style={{ borderColor: `#${label.color}` }}
                      >
                        {label.name}
                      </Badge>
                    ))}
                  </div>
                )}
              </a>
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Son pull request&apos;ler</CardTitle>
            <CardDescription>Açık ve son güncellenen kapalı PR&apos;lar.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {pullRequests.length === 0 && (
              <p className="text-sm text-muted-foreground">PR yok.</p>
            )}
            {pullRequests.map((pr) => (
              <a
                key={pr.number}
                href={pr.html_url}
                target="_blank"
                rel="noreferrer"
                className="block rounded-md border p-3 text-sm hover:bg-muted/40"
              >
                <div className="flex items-start justify-between gap-2">
                  <span className="font-medium">
                    #{pr.number} {pr.title}
                  </span>
                  <Badge
                    variant={
                      pr.merged_at ? "default" : pr.state === "open" ? "secondary" : "destructive"
                    }
                  >
                    {pr.merged_at ? "Merged" : pr.state === "open" ? "Açık" : "Kapalı"}
                  </Badge>
                </div>
                <div className="mt-1 text-xs text-muted-foreground">
                  {pr.user?.login} · <LocalDateTime value={pr.updated_at} />
                </div>
              </a>
            ))}
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
