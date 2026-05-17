import { useCallback, useEffect, useState } from 'react';
import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { StatusBar } from 'expo-status-bar';
const API_URL = 'https://novoceo.api.local.happyapi.fr';
const API_TOKEN = '20298f430f36f8f2a03d102bb74c7483aa14e7a8fc6796242a9de8c21cc32904';

// ---- API helpers ----

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, {
    headers: { Authorization: `Bearer ${API_TOKEN}` },
  });
  if (!res.ok) throw new Error(`API ${res.status}`);
  return res.json() as Promise<T>;
}

async function apiPost(path: string, body: object): Promise<void> {
  await fetch(`${API_URL}${path}`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${API_TOKEN}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });
}

function fmtTime(iso: string): string {
  if (!iso) return '';
  const d = new Date(iso);
  return d.toLocaleString('fr-FR', {
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

// ---- Types ----

interface HealthData {
  load_average: [number, number, number];
  memory_percent: number;
  created_at: string;
}

interface TempData {
  battery: number;
  temperature: number;
  humidity: number;
  created_at: string;
}

interface TempPoint {
  temperature: number;
  created_at: string;
}

interface PriseData {
  state: string;
  created_at: string;
}

interface ContactData {
  contact: boolean | null;
  battery: number | null;
  created_at: string;
}

// ---- Sparkline (native bars) ----

function SparklineFallback({ points }: { points: TempPoint[] }) {
  if (points.length < 2) {
    return (
      <View style={{ marginTop: 8, borderTopWidth: 1, borderTopColor: '#1e293b', paddingTop: 8 }}>
        <Text style={styles.subLabel}>24 dernières heures</Text>
        <Text style={styles.metaText}>aucune donnée</Text>
      </View>
    );
  }

  const temps = points.map((p) => p.temperature);
  const minT = Math.min(...temps);
  const maxT = Math.max(...temps);
  const rangeT = maxT - minT < 0.1 ? 1 : maxT - minT;

  const sampleStep = Math.max(1, Math.floor(points.length / 32));
  const sampled = points.filter((_, i) => i % sampleStep === 0);

  return (
    <View style={{ marginTop: 8, borderTopWidth: 1, borderTopColor: '#1e293b', paddingTop: 8 }}>
      <View style={styles.sparklineRow}>
        <Text style={styles.subLabel}>24 dernières heures</Text>
        <Text style={[styles.metaText, { color: '#92400e' }]}>
          ↓{minT.toFixed(1)}° ↑{maxT.toFixed(1)}°
        </Text>
      </View>
      <View style={{ height: 48, flexDirection: 'row', alignItems: 'flex-end', gap: 1, marginVertical: 4 }}>
        {sampled.map((p, i) => {
          const heightPct = ((p.temperature - minT) / rangeT) * 0.9 + 0.05;
          return (
            <View
              key={i}
              style={{
                flex: 1,
                height: `${heightPct * 100}%`,
                backgroundColor: '#d97706',
                borderRadius: 1,
                opacity: 0.7 + 0.3 * (i / sampled.length),
              }}
            />
          );
        })}
      </View>
      <View style={styles.sparklineRow}>
        <Text style={styles.metaText}>-24h</Text>
        <Text style={styles.metaText}>now</Text>
      </View>
    </View>
  );
}

// ---- Bridge Health Card ----

function CardHealth() {
  const [data, setData] = useState<HealthData | null>(null);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      const d = await apiFetch<HealthData>('/bridge/health');
      setData(d);
      setError(false);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    const id = setInterval(load, 60_000);
    return () => clearInterval(id);
  }, [load]);

  const loads = data?.load_average ?? [0, 0, 0];
  const scale = Math.max(...loads, 0.01);
  const pct = (v: number) => Math.max((v / scale) * 90, 4);
  const mem = data?.memory_percent ?? 0;

  return (
    <View style={[styles.card, styles.cardCyan]}>
      <View style={styles.cardHeader}>
        <View style={[styles.dot, { backgroundColor: '#22d3ee' }]} />
        <Text style={[styles.cardTitle, { color: '#22d3ee' }]}>BRIDGE HEALTH</Text>
        {loading && <ActivityIndicator size="small" color="#22d3ee" style={{ marginLeft: 'auto' }} />}
      </View>

      {error ? (
        <Text style={styles.errorText}>bridge indisponible</Text>
      ) : (
        <View style={{ gap: 4 }}>
          <Text style={styles.subLabel}>load average</Text>
          <View style={{ flexDirection: 'row', gap: 8, height: 72 }}>
            {loads.map((v, i) => (
              <View key={i} style={{ flex: 1, alignItems: 'center', gap: 2 }}>
                <Text style={styles.barValue}>{v.toFixed(2)}</Text>
                <View style={styles.barTrack}>
                  <View
                    style={[
                      styles.barFill,
                      {
                        height: `${pct(v)}%`,
                        backgroundColor: i === 0 ? '#06b6d4' : i === 1 ? '#0891b2' : '#0e7490',
                      },
                    ]}
                  />
                </View>
                <Text style={styles.metaText}>{['1m', '5m', '15m'][i]}</Text>
              </View>
            ))}
          </View>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
            <Text style={styles.subLabel}>mémoire</Text>
            <Text style={[styles.barValue, { color: '#22d3ee' }]}>{mem.toFixed(0)}%</Text>
          </View>
          <View style={styles.memTrack}>
            <View style={[styles.memFill, { width: `${mem}%` }]} />
          </View>
          {data && <Text style={styles.metaText}>{fmtTime(data.created_at)}</Text>}
        </View>
      )}
    </View>
  );
}

// ---- Temperature Card ----

function CardTemperature() {
  const [data, setData] = useState<TempData | null>(null);
  const [history, setHistory] = useState<TempPoint[]>([]);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      const [d, h] = await Promise.all([
        apiFetch<TempData>('/sensor/temperature'),
        apiFetch<TempPoint[]>('/sensor/temperature/history').catch(() => [] as TempPoint[]),
      ]);
      setData(d);
      setHistory(h);
      setError(false);
    } catch {
      setError(true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    const id = setInterval(load, 60_000);
    return () => clearInterval(id);
  }, [load]);

  return (
    <View style={[styles.card, styles.cardAmber]}>
      <View style={styles.cardHeader}>
        <View style={[styles.dot, { backgroundColor: '#fbbf24' }]} />
        <Text style={[styles.cardTitle, { color: '#fbbf24' }]}>TEMPERATURE</Text>
        {loading && <ActivityIndicator size="small" color="#fbbf24" style={{ marginLeft: 'auto' }} />}
      </View>

      {error ? (
        <Text style={styles.errorText}>capteur indisponible</Text>
      ) : data ? (
        <View style={{ gap: 4 }}>
          <View style={{ flexDirection: 'row', alignItems: 'flex-end', gap: 4 }}>
            <Text style={styles.bigTemp}>{data.temperature.toFixed(1)}</Text>
            <Text style={[styles.barValue, { color: '#f59e0b', marginBottom: 6 }]}>°C</Text>
          </View>
          <View style={{ flexDirection: 'row', gap: 24, marginTop: 2 }}>
            <View>
              <Text style={styles.subLabel}>humidité</Text>
              <Text style={[styles.barValue, { color: '#93c5fd' }]}>{data.humidity.toFixed(0)}%</Text>
            </View>
            <View>
              <Text style={styles.subLabel}>batterie</Text>
              <Text style={[styles.barValue, { color: '#86efac' }]}>{data.battery}%</Text>
            </View>
          </View>
          <SparklineFallback points={history} />
          <Text style={[styles.metaText, { marginTop: 2 }]}>{fmtTime(data.created_at)}</Text>
        </View>
      ) : null}
    </View>
  );
}

// ---- Prise Card ----

function CardPrise() {
  const [data, setData] = useState<PriseData | null>(null);
  const [toggling, setToggling] = useState(false);

  const load = useCallback(async () => {
    try {
      const d = await apiFetch<PriseData>('/device/Prise/state');
      setData(d);
    } catch {
      /* silently keep last state */
    }
  }, []);

  useEffect(() => {
    load();
    const id = setInterval(load, 5_000);
    return () => clearInterval(id);
  }, [load]);

  const toggle = async () => {
    if (toggling) return;
    setToggling(true);
    try {
      await apiPost('/device/command', { device: 'Prise', state: 'TOGGLE' });
      await new Promise((r) => setTimeout(r, 1500));
      await load();
    } finally {
      setToggling(false);
    }
  };

  const state = data?.state ?? '?';
  const on = state.toUpperCase() === 'ON';

  return (
    <View style={[styles.card, styles.cardEmer]}>
      <View style={styles.cardHeader}>
        <View style={[styles.dot, { backgroundColor: '#34d399' }]} />
        <Text style={[styles.cardTitle, { color: '#34d399' }]}>PRISE</Text>
      </View>
      <View style={{ gap: 4 }}>
        <View style={{ alignItems: 'center', marginVertical: 8 }}>
          <View style={[styles.circle, on ? styles.glowOn : styles.glowOff]}>
            <Text style={[styles.circleText, { color: on ? '#34d399' : '#a78bfa' }]}>{state}</Text>
          </View>
          <View
            style={[
              styles.circleDot,
              { backgroundColor: on ? '#34d399' : '#7c3aed' },
            ]}
          />
        </View>
        {data && <Text style={[styles.metaText, { textAlign: 'center' }]}>{fmtTime(data.created_at)}</Text>}
        <Pressable
          onPress={toggle}
          disabled={toggling}
          style={({ pressed }) => [styles.toggleBtn, pressed && { opacity: 0.7 }]}
        >
          {toggling ? (
            <ActivityIndicator size="small" color="#cbd5e1" />
          ) : (
            <Text style={styles.toggleBtnText}>Toggle</Text>
          )}
        </Pressable>
      </View>
    </View>
  );
}

// ---- Contact/Door Card ----

function CardContact() {
  const [data, setData] = useState<ContactData | null>(null);
  const [loading, setLoading] = useState(true);

  const load = useCallback(async () => {
    try {
      const d = await apiFetch<ContactData>('/device/Contacteur%20porte/contact');
      setData(d);
    } catch {
      /* silently keep last state */
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
    const id = setInterval(load, 5_000);
    return () => clearInterval(id);
  }, [load]);

  const contact = data?.contact;
  const label = contact == null ? '?' : contact ? 'FERME' : 'OUVERT';
  const color = contact == null ? '#64748b' : contact ? '#c4b5fd' : '#f87171';
  const dotColor = contact == null ? '#475569' : contact ? '#a78bfa' : '#ef4444';
  const glowStyle = contact == null ? {} : contact ? styles.glowOff : styles.glowAlert;

  return (
    <View style={[styles.card, styles.cardViolet]}>
      <View style={styles.cardHeader}>
        <View style={[styles.dot, { backgroundColor: '#a78bfa' }]} />
        <Text style={[styles.cardTitle, { color: '#a78bfa' }]}>PORTE</Text>
        {loading && <ActivityIndicator size="small" color="#a78bfa" style={{ marginLeft: 'auto' }} />}
      </View>
      <View style={{ gap: 4 }}>
        <View style={{ alignItems: 'center', marginVertical: 8 }}>
          <View style={[styles.circle, glowStyle]}>
            <Text style={[styles.circleText, { color, fontSize: 14 }]}>{label}</Text>
          </View>
          <View style={[styles.circleDot, { backgroundColor: dotColor }]} />
        </View>
        {data?.battery != null && (
          <View style={{ marginTop: 4, alignItems: 'center' }}>
            <Text style={styles.subLabel}>batterie</Text>
            <Text style={[styles.barValue, { color: '#86efac' }]}>{data.battery}%</Text>
          </View>
        )}
        {data && <Text style={[styles.metaText, { textAlign: 'center', marginTop: 2 }]}>{fmtTime(data.created_at)}</Text>}
      </View>
    </View>
  );
}

// ---- Root ----

export default function App() {
  const [refreshing, setRefreshing] = useState(false);
  const [key, setKey] = useState(0);

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    setKey((k) => k + 1);
    setTimeout(() => setRefreshing(false), 600);
  }, []);

  return (
    <SafeAreaView style={styles.safeArea}>
      <StatusBar style="light" />
      <ScrollView
        contentContainerStyle={styles.container}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor="#2563eb" colors={['#2563eb']} />
        }
      >
        {/* Header */}
        <View style={styles.header}>
          <View>
            <View style={{ flexDirection: 'row', alignItems: 'baseline' }}>
              <Text style={styles.logoNovo}>NOVO</Text>
              <Text style={[styles.logoNovo, { color: '#2563eb' }]}>CEO</Text>
            </View>
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6, marginTop: 3 }}>
              <View style={styles.logoDash} />
              <Text style={styles.logoSub}>GROUPE</Text>
              <View style={styles.logoDash} />
            </View>
          </View>
          <View style={styles.statusDot} />
        </View>

        {/* Cards */}
        <View key={key} style={styles.grid}>
          <CardHealth />
          <CardTemperature />
          <CardPrise />
          <CardContact />
        </View>

        <Text style={styles.footer}>novoceo · zigbee2mqtt monitoring</Text>
      </ScrollView>
    </SafeAreaView>
  );
}

// ---- Styles ----

const C = {
  bg: '#020617',
  card: '#0f172a',
  slate700: '#334155',
  slate800: '#1e293b',
};

const styles = StyleSheet.create({
  safeArea: { flex: 1, backgroundColor: C.bg },
  container: { padding: 16, paddingBottom: 48 },

  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 24,
  },
  logoNovo: {
    fontWeight: '700',
    fontSize: 20,
    color: '#f1f5f9',
    letterSpacing: -0.8,
  },
  logoSub: { fontSize: 8, color: '#475569', letterSpacing: 5 },
  logoDash: { height: 1, width: 16, backgroundColor: '#2563eb' },
  statusDot: { height: 8, width: 8, borderRadius: 4, backgroundColor: '#34d399' },

  grid: { gap: 12 },

  card: {
    backgroundColor: C.card,
    borderRadius: 16,
    padding: 16,
    borderWidth: 1,
  },
  cardCyan: { borderColor: 'rgba(6,182,212,0.15)' },
  cardAmber: { borderColor: 'rgba(245,158,11,0.15)' },
  cardEmer: { borderColor: 'rgba(52,211,153,0.15)' },
  cardViolet: { borderColor: 'rgba(139,92,246,0.15)' },

  cardHeader: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 14 },
  dot: { height: 7, width: 7, borderRadius: 3.5 },
  cardTitle: { fontSize: 10, fontWeight: '500', letterSpacing: 2 },

  subLabel: { fontSize: 10, color: '#64748b', marginBottom: 2 },
  barValue: { fontSize: 13, color: '#cbd5e1' },
  metaText: { fontSize: 10, color: '#334155' },
  errorText: { fontSize: 11, color: '#f87171' },

  barTrack: {
    flex: 1,
    width: '100%',
    backgroundColor: C.slate800,
    borderRadius: 2,
    overflow: 'hidden',
    justifyContent: 'flex-end',
  },
  barFill: { width: '100%', borderRadius: 2 },
  memTrack: { height: 6, backgroundColor: C.slate800, borderRadius: 3, overflow: 'hidden', marginTop: 2 },
  memFill: { height: '100%', backgroundColor: '#0891b2', borderRadius: 3 },

  bigTemp: { fontSize: 48, fontWeight: '500', color: '#fcd34d', lineHeight: 56 },

  sparklineRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },

  circle: {
    height: 96,
    width: 96,
    borderRadius: 48,
    backgroundColor: C.slate800,
    borderWidth: 2,
    borderColor: C.slate700,
    alignItems: 'center',
    justifyContent: 'center',
  },
  circleDot: {
    position: 'absolute',
    bottom: 8,
    right: '50%',
    marginRight: -52,
    height: 14,
    width: 14,
    borderRadius: 7,
    borderWidth: 2,
    borderColor: C.card,
    marginTop: -14,
  },
  circleText: { fontSize: 18, fontWeight: '500' },
  glowOn: { borderColor: 'rgba(52,211,153,0.35)', shadowColor: '#34d399', shadowOpacity: 0.5, shadowRadius: 12, elevation: 6 },
  glowOff: { borderColor: 'rgba(139,92,246,0.25)', shadowColor: '#8b5cf6', shadowOpacity: 0.3, shadowRadius: 10, elevation: 4 },
  glowAlert: { borderColor: 'rgba(239,68,68,0.35)', shadowColor: '#ef4444', shadowOpacity: 0.5, shadowRadius: 12, elevation: 6 },

  toggleBtn: {
    marginTop: 12,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: C.slate800,
    borderWidth: 1,
    borderColor: C.slate700,
    alignItems: 'center',
  },
  toggleBtnText: { fontSize: 13, fontWeight: '500', color: '#cbd5e1' },

  footer: { marginTop: 32, textAlign: 'center', fontSize: 11, color: '#1e3a5f' },
});
