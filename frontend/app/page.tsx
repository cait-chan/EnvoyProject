import Link from "next/link";

const BACKEND_URL = process.env.BACKEND_URL ?? "http://localhost:8080";

async function getStats() {
  try {
    const res = await fetch(`${BACKEND_URL}/api/stats`, { cache: "no-store" });
    if (!res.ok) throw new Error("Failed to fetch stats");
    return res.json();
  } catch {
    return null;
  }
}

async function getClusters() {
  try {
    const res = await fetch(`${BACKEND_URL}/api/clusters`, {
      cache: "no-store",
    });
    if (!res.ok) throw new Error("Failed to fetch clusters");
    return res.json();
  } catch {
    return null;
  }
}

type StatEntry = { name: string; value?: number };

/** Envoy JSON sometimes uses number, string, or typed metrics arrays. */
function statNum(item: Record<string, unknown>): number {
  const v = item.value;
  if (typeof v === "number" && !Number.isNaN(v)) return v;
  if (typeof v === "string") {
    const n = Number(v);
    return Number.isFinite(n) ? n : 0;
  }
  const metrics = item.metrics;
  if (Array.isArray(metrics)) {
    for (const m of metrics) {
      if (
        m &&
        typeof m === "object" &&
        "value" in m &&
        typeof (m as { value: unknown }).value === "number"
      ) {
        return (m as { value: number }).value;
      }
    }
  }
  return 0;
}

function parseStats(stats: unknown) {
  const out = {
    totalRequests: 0,
    activeConnections: 0,
    clusterRequests: {} as Record<string, number>,
    clusterConnections: {} as Record<string, number>,
    clusterUpstreamActive: {} as Record<string, number>,
  };
  if (!stats || typeof stats !== "object" || !("stats" in stats)) return out;
  const list = (stats as { stats?: StatEntry[] }).stats;
  if (!Array.isArray(list)) return out;

  for (const item of list) {
    if (!item || typeof item.name !== "string") continue;
    const name = item.name;
    const raw = item as unknown as Record<string, unknown>;
    const value = statNum(raw);
    // Sum HCM request totals for ingress listeners (exclude http.admin = :9901).
    const httpRqTotal = name.match(/^http\.([^.]+)\.downstream_rq_total$/);
    if (httpRqTotal && httpRqTotal[1] !== "admin") {
      out.totalRequests += value;
    }
    // Active *client→Envoy* connections: use listener gauges (excludes :9901 admin listener).
    // http.*.downstream_cx_active can be missing or 0 in JSON depending on scope; listeners match UX.
    if (
      name.startsWith("listener.") &&
      name.endsWith(".downstream_cx_active") &&
      !name.includes("_9901")
    ) {
      out.activeConnections += value;
    }
    const m = name.match(
      /^cluster\.([^.]+)\.(upstream_rq_total|upstream_cx_total|upstream_cx_active)$/,
    );
    if (m) {
      const [, cluster, kind] = m;
      if (kind === "upstream_rq_total") out.clusterRequests[cluster] = value;
      else if (kind === "upstream_cx_total")
        out.clusterConnections[cluster] = value;
      else out.clusterUpstreamActive[cluster] = value;
    }
  }
  return out;
}

function parseClusters(
  clusters: unknown,
): { name: string; healthy: number; total: number }[] {
  if (
    !clusters ||
    typeof clusters !== "object" ||
    !("cluster_statuses" in clusters)
  )
    return [];
  const statuses = (clusters as { cluster_statuses?: unknown[] })
    .cluster_statuses;
  if (!Array.isArray(statuses)) return [];

  return statuses.map((s: unknown) => {
    const o = s as Record<string, unknown>;
    const name = typeof o.name === "string" ? o.name : String(o.name ?? "");
    const hostStatus =
      (o.host_statuses as Array<Record<string, unknown>>) ?? [];
    const total = hostStatus.length;
    const healthy = total > 0 ? total : 0;
    return { name, healthy, total };
  });
}

export default async function DashboardPage() {
  const [statsJson, clustersJson] = await Promise.all([
    getStats(),
    getClusters(),
  ]);
  const metrics = statsJson ? parseStats(statsJson) : null;
  const clusterList = clustersJson ? parseClusters(clustersJson) : [];
  const clusterNames = metrics
    ? [
        ...new Set([
          ...Object.keys(metrics.clusterRequests),
          ...Object.keys(metrics.clusterConnections),
          ...Object.keys(metrics.clusterUpstreamActive),
        ]),
      ]
    : [];

  return (
    <div className="min-h-screen bg-zinc-50 p-8 dark:bg-zinc-950">
      <div className="mx-auto max-w-4xl">
        <div className="mb-8 flex items-center justify-between">
          <h1 className="text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
            Envoy Metrics Dashboard
          </h1>
          <Link
            href="/"
            className="text-sm text-zinc-600 hover:underline dark:text-zinc-400"
          >
            ← Home
          </Link>
        </div>

        {!statsJson && !clustersJson && (
          <p className="rounded-lg bg-amber-100 p-4 text-amber-900 dark:bg-amber-900/30 dark:text-amber-200">
            Could not load data. Is the backend on :8080 and Envoy on :9901?
          </p>
        )}

        {metrics && (
          <section className="mb-8">
            <h2 className="mb-4 text-lg font-medium text-zinc-800 dark:text-zinc-200">
              Key metrics
            </h2>
            <div className="grid gap-4 sm:grid-cols-2">
              <div className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">
                  Total requests (through proxy)
                </p>
                <p className="mt-1 text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
                  {metrics.totalRequests}
                </p>
              </div>
              <div className="rounded-lg border border-zinc-200 bg-white p-4 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
                <p className="text-sm text-zinc-500 dark:text-zinc-400">
                  Active downstream connections
                </p>
                <p className="mt-1 text-2xl font-semibold text-zinc-900 dark:text-zinc-100">
                  {metrics.activeConnections}
                </p>
              </div>
            </div>
          </section>
        )}

        {clusterNames.length > 0 && metrics && (
          <section className="mb-8">
            <h2 className="mb-4 text-lg font-medium text-zinc-800 dark:text-zinc-200">
              Traffic per cluster
            </h2>
            <div className="overflow-hidden rounded-lg border border-zinc-200 bg-white dark:border-zinc-800 dark:bg-zinc-900">
              <table className="w-full table-fixed text-left text-sm">
                <colgroup>
                  <col className="w-[40%]" />
                  <col className="w-[20%]" />
                  <col className="w-[20%]" />
                  <col className="w-[20%]" />
                </colgroup>
                <thead>
                  <tr className="border-b border-zinc-200 bg-zinc-50 dark:border-zinc-800 dark:bg-zinc-800/50">
                    <th className="min-w-0 px-4 py-3 font-medium text-zinc-700 dark:text-zinc-300">
                      Cluster
                    </th>
                    <th className="px-4 py-3 text-center font-medium text-zinc-700 dark:text-zinc-300">
                      Requests
                    </th>
                    <th className="px-4 py-3 text-center font-medium text-zinc-700 dark:text-zinc-300">
                      Upstream opened (total)
                    </th>
                    <th className="px-4 py-3 text-center font-medium text-zinc-700 dark:text-zinc-300">
                      Upstream active
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {clusterNames.map((name) => (
                    <tr
                      key={name}
                      className="border-b border-zinc-100 last:border-0 dark:border-zinc-800"
                    >
                      <td className="min-w-0 break-all px-4 py-3 font-mono text-sm text-zinc-900 dark:text-zinc-100">
                        {name}
                      </td>
                      <td className="px-4 py-3 text-center tabular-nums text-zinc-600 dark:text-zinc-400">
                        {metrics.clusterRequests[name] ?? "—"}
                      </td>
                      <td className="px-4 py-3 text-center tabular-nums text-zinc-600 dark:text-zinc-400">
                        {metrics.clusterConnections[name] ?? "—"}
                      </td>
                      <td className="px-4 py-3 text-center tabular-nums text-zinc-600 dark:text-zinc-400">
                        {metrics.clusterUpstreamActive[name] ?? "—"}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </section>
        )}

        {clusterList.length > 0 && (
          <section className="mb-8">
            <h2 className="mb-4 text-lg font-medium text-zinc-800 dark:text-zinc-200">
              Cluster health
            </h2>
            <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {clusterList.map((c) => (
                <div
                  key={c.name}
                  className="rounded-lg border border-zinc-200 bg-white p-4 dark:border-zinc-800 dark:bg-zinc-900"
                >
                  <p className="font-mono font-medium text-zinc-900 dark:text-zinc-100">
                    {c.name}
                  </p>
                  <p className="mt-1 text-sm text-zinc-500 dark:text-zinc-400">
                    {c.healthy}/{c.total} healthy
                  </p>
                </div>
              ))}
            </div>
          </section>
        )}

        <details className="mt-8">
          <summary className="cursor-pointer text-sm font-medium text-zinc-500 dark:text-zinc-400">
            View raw JSON (debug)
          </summary>
          <div className="mt-4 space-y-6">
            <div>
              <h3 className="mb-2 text-sm text-zinc-600 dark:text-zinc-400">
                Stats
              </h3>
              <pre className="max-h-64 overflow-auto rounded-lg bg-zinc-100 p-4 text-xs dark:bg-zinc-900">
                {statsJson ? JSON.stringify(statsJson, null, 2) : "No data"}
              </pre>
            </div>
            <div>
              <h3 className="mb-2 text-sm text-zinc-600 dark:text-zinc-400">
                Clusters
              </h3>
              <pre className="max-h-64 overflow-auto rounded-lg bg-zinc-100 p-4 text-xs dark:bg-zinc-900">
                {clustersJson
                  ? JSON.stringify(clustersJson, null, 2)
                  : "No data"}
              </pre>
            </div>
          </div>
        </details>
      </div>
    </div>
  );
}
