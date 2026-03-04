import { useMemo } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type {
  TabDataPoint,
  MemoryDataPoint,
  ServerDataPoint,
} from "../../stores/useAppStore";

interface Props {
  data: TabDataPoint[];
  memoryData?: MemoryDataPoint[];
  serverData?: ServerDataPoint[];
  instances: { id: string; profileName: string }[];
  selectedInstanceId: string | null;
  onSelectInstance: (id: string) => void;
}

// Colors for different instances
const COLORS = [
  "#f97316", // orange (primary)
  "#3b82f6", // blue
  "#22c55e", // green
  "#eab308", // yellow
  "#ef4444", // red
  "#8b5cf6", // purple
  "#ec4899", // pink
  "#14b8a6", // teal
];

function formatTime(timestamp: number): string {
  return new Date(timestamp).toLocaleTimeString("en-GB", {
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function TabsChart({
  data,
  memoryData,
  serverData,
  instances,
  selectedInstanceId,
  onSelectInstance,
}: Props) {
  const instanceColors = useMemo(() => {
    const colors: Record<string, string> = {};
    instances.forEach((inst, i) => {
      colors[inst.id] = COLORS[i % COLORS.length];
    });
    return colors;
  }, [instances]);

  // Merge tab, memory, and server data by timestamp
  const mergedData = useMemo(() => {
    const memByTime = new Map((memoryData || []).map((m) => [m.timestamp, m]));
    const serverByTime = new Map(
      (serverData || []).map((s) => [s.timestamp, s]),
    );

    // Use tab data as base, or server data if no tabs
    const baseData =
      data.length > 0
        ? data
        : (serverData || []).map((s) => ({ timestamp: s.timestamp }));

    return baseData.map((d) => {
      const merged: Record<string, number> = { timestamp: d.timestamp };

      // Add tab data if present
      for (const [key, val] of Object.entries(d)) {
        if (key !== "timestamp") {
          merged[key] = val as number;
        }
      }

      // Add memory keys with _mem suffix
      const mem = memByTime.get(d.timestamp);
      if (mem) {
        for (const [key, val] of Object.entries(mem)) {
          if (key !== "timestamp") {
            merged[`${key}_mem`] = val;
          }
        }
      }

      // Add server metrics
      const srv = serverByTime.get(d.timestamp);
      if (srv) {
        merged.goHeapMB = srv.goHeapMB;
      }

      return merged;
    });
  }, [data, memoryData, serverData]);

  // Show empty state if no data or too few points to render meaningfully
  if (mergedData.length < 2) {
    return (
      <div className="flex h-50 items-center justify-center rounded-lg border border-border-subtle bg-bg-surface text-sm text-text-muted">
        {mergedData.length === 0
          ? "Collecting data..."
          : "Waiting for more data..."}
      </div>
    );
  }

  const hasMemory = memoryData && memoryData.length > 0;
  const hasServer = serverData && serverData.length > 0;

  return (
    <div className="rounded-lg border border-border-subtle bg-bg-surface">
      <ResponsiveContainer width="100%" height={200}>
        <LineChart
          data={mergedData}
          margin={{
            top: 16,
            right: hasMemory || hasServer ? 50 : 16,
            bottom: 8,
            left: 8,
          }}
        >
          <XAxis
            dataKey="timestamp"
            tickFormatter={formatTime}
            stroke="#666"
            fontSize={11}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            yAxisId="tabs"
            stroke="#666"
            fontSize={11}
            allowDecimals={false}
            domain={[0, "auto"]}
            tickLine={false}
            axisLine={false}
            width={30}
          />
          {(hasMemory || hasServer) && (
            <YAxis
              yAxisId="memory"
              orientation="right"
              stroke="#888"
              fontSize={11}
              allowDecimals={false}
              domain={[0, "auto"]}
              tickLine={false}
              axisLine={false}
              width={40}
              tickFormatter={(v) => `${v}MB`}
            />
          )}
          <Tooltip
            contentStyle={{
              background: "#1a1a1a",
              border: "1px solid #333",
              borderRadius: "6px",
              fontSize: "12px",
            }}
            labelFormatter={(label) => formatTime(label as number)}
            formatter={(value, name) => {
              const nameStr = String(name);
              if (nameStr === "goHeapMB") {
                return [`${value}MB`, "Server Heap"];
              }
              const isMemory = nameStr.endsWith("_mem");
              const instId = isMemory ? nameStr.replace("_mem", "") : nameStr;
              const inst = instances.find((i) => i.id === instId);
              const label = inst?.profileName || instId;
              return [
                isMemory ? `${value}MB` : value,
                isMemory ? `${label} (mem)` : `${label} (tabs)`,
              ];
            }}
          />
          {/* Tab count lines (solid) */}
          {instances.map((inst) => (
            <Line
              key={inst.id}
              yAxisId="tabs"
              type="monotone"
              dataKey={inst.id}
              name={inst.id}
              stroke={instanceColors[inst.id]}
              strokeWidth={selectedInstanceId === inst.id ? 3 : 1.5}
              strokeOpacity={
                selectedInstanceId && selectedInstanceId !== inst.id ? 0.3 : 1
              }
              dot={false}
              activeDot={{
                r: 4,
                onClick: () => onSelectInstance(inst.id),
                style: { cursor: "pointer" },
              }}
            />
          ))}
          {/* Memory lines (dashed) */}
          {hasMemory &&
            instances.map((inst) => (
              <Line
                key={`${inst.id}_mem`}
                yAxisId="memory"
                type="monotone"
                dataKey={`${inst.id}_mem`}
                name={`${inst.id}_mem`}
                stroke={instanceColors[inst.id]}
                strokeWidth={selectedInstanceId === inst.id ? 2 : 1}
                strokeOpacity={
                  selectedInstanceId && selectedInstanceId !== inst.id
                    ? 0.2
                    : 0.6
                }
                strokeDasharray="4 2"
                dot={false}
              />
            ))}
          {/* Server heap line (dotted, gray) */}
          {hasServer && (
            <Line
              yAxisId="memory"
              type="monotone"
              dataKey="goHeapMB"
              name="goHeapMB"
              stroke="#888"
              strokeWidth={1.5}
              strokeDasharray="2 2"
              dot={false}
            />
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
