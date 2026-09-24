// A single-hue magnitude bar — for comparing counts across categories
// of the same measure (age range, mood preference, ...), not identity
// across series, so one hue (the design system's primary) is correct
// here rather than a categorical palette. Thin (8px), rounded ends, on
// a muted track — see the dataviz skill's mark spec.
export function DistributionBar({
  label,
  count,
  total,
}: {
  label: string;
  count: number;
  total: number;
}) {
  const pct = total > 0 ? (count / total) * 100 : 0;
  return (
    <div className="space-y-1">
      <div className="flex items-center justify-between text-sm">
        <span>{label}</span>
        <span className="text-muted-foreground">
          {count} · %{Math.round(pct)}
        </span>
      </div>
      <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
        <div
          className="h-full rounded-full bg-primary"
          style={{ width: `${pct}%` }}
        />
      </div>
    </div>
  );
}
