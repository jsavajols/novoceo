import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  ActivityIndicator,
  Pressable,
  RefreshControl,
  SafeAreaView,
  ScrollView,
  StyleSheet,
  Text,
  useColorScheme,
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

// ---- Theme ----

type ThemeMode = 'dark' | 'light' | 'system';

interface Palette {
  bg: string;
  card: string;
  cardCyan: string;
  cardAmber: string;
  cardEmer: string;
  cardViolet: string;
  borderCyan: string;
  borderAmber: string;
  borderEmer: string;
  borderViolet: string;
  track: string;
  text: string;
  textMeta: string;
  label: string;
  circleBg: string;
  circleBorder: string;
  btnBg: string;
  btnBorder: string;
  logoColor: string;
  logoSub: string;
  footer: string;
  divider: string;
  bigTemp: string;
  accentHumidity: string;
  accentBattery: string;
  barValue: string;
  pillBg: string;
  pillBorder: string;
  pillActiveBg: string;
  pillText: string;
  pillActiveText: string;
  accentCyan: string;
  accentAmber: string;
  accentEmer: string;
  accentViolet: string;
  stateOn: string;
  stateOff: string;
  contactClosed: string;
  contactOpen: string;
}

const DARK: Palette = {
  bg: '#020617',
  card: '#0f172a',
  cardCyan: '#0f172a',
  cardAmber: '#0f172a',
  cardEmer: '#0f172a',
  cardViolet: '#0f172a',
  borderCyan: 'rgba(6,182,212,0.18)',
  borderAmber: 'rgba(245,158,11,0.18)',
  borderEmer: 'rgba(52,211,153,0.18)',
  borderViolet: 'rgba(139,92,246,0.18)',
  track: '#1e293b',
  text: '#cbd5e1',
  textMeta: '#475569',
  label: '#64748b',
  circleBg: '#1e293b',
  circleBorder: '#334155',
  btnBg: '#1e293b',
  btnBorder: '#334155',
  logoColor: '#f1f5f9',
  logoSub: '#475569',
  footer: '#334155',
  divider: '#1e293b',
  bigTemp: '#fcd34d',
  accentHumidity: '#93c5fd',
  accentBattery: '#86efac',
  barValue: '#cbd5e1',
  pillBg: '#1e293b',
  pillBorder: '#334155',
  pillActiveBg: '#2563eb',
  pillText: '#475569',
  pillActiveText: '#f1f5f9',
  accentCyan: '#22d3ee',
  accentAmber: '#fbbf24',
  accentEmer: '#34d399',
  accentViolet: '#a78bfa',
  stateOn: '#34d399',
  stateOff: '#a78bfa',
  contactClosed: '#c4b5fd',
  contactOpen: '#f87171',
};

const LIGHT: Palette = {
  bg: '#f1f5f9',
  card: '#ffffff',
  cardCyan: '#f0fdff',
  cardAmber: '#fffcf0',
  cardEmer: '#f0fdf4',
  cardViolet: '#faf5ff',
  borderCyan: 'rgba(6,182,212,0.3)',
  borderAmber: 'rgba(245,158,11,0.3)',
  borderEmer: 'rgba(52,211,153,0.3)',
  borderViolet: 'rgba(139,92,246,0.3)',
  track: '#e2e8f0',
  text: '#334155',
  textMeta: '#64748b',
  label: '#475569',
  circleBg: '#f8fafc',
  circleBorder: '#cbd5e1',
  btnBg: '#f1f5f9',
  btnBorder: '#cbd5e1',
  logoColor: '#0f172a',
  logoSub: '#94a3b8',
  footer: '#94a3b8',
  divider: '#e2e8f0',
  bigTemp: '#d97706',
  accentHumidity: '#2563eb',
  accentBattery: '#16a34a',
  barValue: '#1e293b',
  pillBg: '#e2e8f0',
  pillBorder: '#cbd5e1',
  pillActiveBg: '#2563eb',
  pillText: '#64748b',
  pillActiveText: '#ffffff',
  accentCyan: '#0891b2',
  accentAmber: '#d97706',
  accentEmer: '#059669',
  accentViolet: '#7c3aed',
  stateOn: '#059669',
  stateOff: '#7c3aed',
  contactClosed: '#7c3aed',
  contactOpen: '#dc2626',
};

type AppStyles = ReturnType<typeof makeStyles>;

function makeStyles(c: Palette) {
  return StyleSheet.create({
    safeArea: { flex: 1, backgroundColor: c.bg },
    container: { padding: 16, paddingBottom: 48 },

    header: {
      flexDirection: 'row',
      justifyContent: 'space-between',
      alignItems: 'center',
      marginBottom: 24,
    },
    logoNovo: { fontWeight: '700', fontSize: 20, color: c.logoColor, letterSpacing: -0.8 },
    logoSub: { fontSize: 8, color: c.logoSub, letterSpacing: 5 },
    logoDash: { height: 1, width: 16, backgroundColor: '#2563eb' },
    statusDot: { height: 8, width: 8, borderRadius: 4, backgroundColor: '#34d399' },

    headerRight: { flexDirection: 'row', alignItems: 'center', gap: 10 },
    themePill: {
      width: 28,
      height: 28,
      borderRadius: 14,
      backgroundColor: c.pillBg,
      borderWidth: 1,
      borderColor: c.pillBorder,
      alignItems: 'center',
      justifyContent: 'center',
    },
    themePillActive: { backgroundColor: c.pillActiveBg, borderColor: c.pillActiveBg },
    themePillText: { fontSize: 13, color: c.pillText },
    themePillTextActive: { color: c.pillActiveText },

    grid: { gap: 16 },

    card: { borderRadius: 16, padding: 20, borderWidth: 1 },
    cardHeader: { flexDirection: 'row', alignItems: 'center', gap: 8, marginBottom: 16 },
    dot: { height: 7, width: 7, borderRadius: 3.5 },
    cardTitle: { fontSize: 11, fontWeight: '500', letterSpacing: 2 },

    subLabel: { fontSize: 14, color: c.label, marginBottom: 3 },
    barValue: { fontSize: 15, color: c.barValue },
    metaText: { fontSize: 12, color: c.textMeta },
    errorText: { fontSize: 12, color: '#f87171' },

    barTrack: {
      flex: 1,
      width: '100%',
      backgroundColor: c.track,
      borderRadius: 2,
      overflow: 'hidden',
      justifyContent: 'flex-end',
    },
    barFill: { width: '100%', borderRadius: 2 },
    memTrack: { height: 6, backgroundColor: c.track, borderRadius: 3, overflow: 'hidden', marginTop: 2 },
    memFill: { height: '100%', backgroundColor: '#0891b2', borderRadius: 3 },

    bigTemp: { fontSize: 48, fontWeight: '500', color: c.bigTemp, lineHeight: 56 },

    sparklineRow: { flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginBottom: 4 },

    circle: {
      height: 96,
      width: 96,
      borderRadius: 48,
      backgroundColor: c.circleBg,
      borderWidth: 2,
      borderColor: c.circleBorder,
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
      borderColor: c.card,
      marginTop: -14,
    },
    circleText: { fontSize: 20, fontWeight: '500' },
    glowOn: {
      borderColor: 'rgba(52,211,153,0.35)',
      shadowColor: '#34d399',
      shadowOpacity: 0.5,
      shadowRadius: 12,
      elevation: 6,
    },
    glowOff: {
      borderColor: 'rgba(139,92,246,0.25)',
      shadowColor: '#8b5cf6',
      shadowOpacity: 0.3,
      shadowRadius: 10,
      elevation: 4,
    },
    glowAlert: {
      borderColor: 'rgba(239,68,68,0.35)',
      shadowColor: '#ef4444',
      shadowOpacity: 0.5,
      shadowRadius: 12,
      elevation: 6,
    },

    toggleBtn: {
      marginTop: 12,
      paddingVertical: 10,
      borderRadius: 12,
      backgroundColor: c.btnBg,
      borderWidth: 1,
      borderColor: c.btnBorder,
      alignItems: 'center',
    },
    toggleBtnText: { fontSize: 15, fontWeight: '500', color: c.text },

    footer: { marginTop: 32, textAlign: 'center', fontSize: 12, color: c.footer },
  });
}

// ---- Theme context ----

interface ThemeCtx {
  mode: ThemeMode;
  setMode: (m: ThemeMode) => void;
  colors: Palette;
  dark: boolean;
  s: AppStyles;
}

const ThemeContext = createContext<ThemeCtx>({
  mode: 'system',
  setMode: () => {},
  colors: DARK,
  dark: true,
  s: makeStyles(DARK),
});

function useTheme() {
  return useContext(ThemeContext);
}

function ThemeProvider({ children }: { children: React.ReactNode }) {
  const sys = useColorScheme();
  const [mode, setMode] = useState<ThemeMode>('system');
  const dark = mode === 'system' ? sys === 'dark' : mode === 'dark';
  const colors = dark ? DARK : LIGHT;
  const s = useMemo(() => makeStyles(colors), [colors]);
  return (
    <ThemeContext.Provider value={{ mode, setMode, colors, dark, s }}>
      {children}
    </ThemeContext.Provider>
  );
}

// ---- Theme selector ----

function ThemeSelector() {
  const { mode, setMode, s } = useTheme();
  const opts: { m: ThemeMode; icon: string }[] = [
    { m: 'light', icon: '☀' },
    { m: 'system', icon: '◑' },
    { m: 'dark', icon: '☾' },
  ];
  return (
    <View style={{ flexDirection: 'row', gap: 6, alignItems: 'center' }}>
      {opts.map(({ m, icon }) => (
        <Pressable
          key={m}
          onPress={() => setMode(m)}
          style={[s.themePill, mode === m && s.themePillActive]}
        >
          <Text style={[s.themePillText, mode === m && s.themePillTextActive]}>{icon}</Text>
        </Pressable>
      ))}
    </View>
  );
}

// ---- Sparkline (native bars) ----

function SparklineFallback({ points }: { points: TempPoint[] }) {
  const { s, colors } = useTheme();

  if (points.length < 2) {
    return (
      <View style={{ marginTop: 8, borderTopWidth: 1, borderTopColor: colors.divider, paddingTop: 8 }}>
        <Text style={s.subLabel}>24 dernières heures</Text>
        <Text style={s.metaText}>aucune donnée</Text>
      </View>
    );
  }

  const temps = points.map((pt) => pt.temperature);
  const minT = Math.min(...temps);
  const maxT = Math.max(...temps);
  const rangeT = maxT - minT < 0.1 ? 1 : maxT - minT;

  const sampleStep = Math.max(1, Math.floor(points.length / 32));
  const sampled = points.filter((_, i) => i % sampleStep === 0);

  return (
    <View style={{ marginTop: 8, borderTopWidth: 1, borderTopColor: colors.divider, paddingTop: 8 }}>
      <View style={s.sparklineRow}>
        <Text style={s.subLabel}>24 dernières heures</Text>
        <Text style={[s.metaText, { color: '#92400e' }]}>
          ↓{minT.toFixed(1)}° ↑{maxT.toFixed(1)}°
        </Text>
      </View>
      <View style={{ height: 48, flexDirection: 'row', alignItems: 'flex-end', gap: 1, marginVertical: 4 }}>
        {sampled.map((pt, i) => {
          const heightPct = ((pt.temperature - minT) / rangeT) * 0.9 + 0.05;
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
      <View style={s.sparklineRow}>
        <Text style={s.metaText}>-24h</Text>
        <Text style={s.metaText}>now</Text>
      </View>
    </View>
  );
}

// ---- Bridge Health Card ----

function CardHealth() {
  const [data, setData] = useState<HealthData | null>(null);
  const [error, setError] = useState(false);
  const [loading, setLoading] = useState(true);
  const { s, colors } = useTheme();

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
    <View style={[s.card, { backgroundColor: colors.cardCyan, borderColor: colors.borderCyan }]}>
      <View style={s.cardHeader}>
        <View style={[s.dot, { backgroundColor: colors.accentCyan }]} />
        <Text style={[s.cardTitle, { color: colors.accentCyan }]}>BRIDGE HEALTH</Text>
        {loading && <ActivityIndicator size="small" color={colors.accentCyan} style={{ marginLeft: 'auto' }} />}
      </View>

      {error ? (
        <Text style={s.errorText}>bridge indisponible</Text>
      ) : (
        <View style={{ gap: 4 }}>
          <Text style={s.subLabel}>load average</Text>
          <View style={{ flexDirection: 'row', gap: 8, height: 72 }}>
            {loads.map((v, i) => (
              <View key={i} style={{ flex: 1, alignItems: 'center', gap: 2 }}>
                <Text style={s.barValue}>{v.toFixed(2)}</Text>
                <View style={s.barTrack}>
                  <View
                    style={[
                      s.barFill,
                      {
                        height: `${pct(v)}%`,
                        backgroundColor: i === 0 ? '#06b6d4' : i === 1 ? '#0891b2' : '#0e7490',
                      },
                    ]}
                  />
                </View>
                <Text style={s.metaText}>{['1m', '5m', '15m'][i]}</Text>
              </View>
            ))}
          </View>
          <View style={{ flexDirection: 'row', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
            <Text style={s.subLabel}>mémoire</Text>
            <Text style={[s.barValue, { color: colors.accentCyan }]}>{mem.toFixed(0)}%</Text>
          </View>
          <View style={s.memTrack}>
            <View style={[s.memFill, { width: `${mem}%` }]} />
          </View>
          {data && <Text style={s.metaText}>{fmtTime(data.created_at)}</Text>}
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
  const { s, colors } = useTheme();

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
    <View style={[s.card, { backgroundColor: colors.cardAmber, borderColor: colors.borderAmber }]}>
      <View style={s.cardHeader}>
        <View style={[s.dot, { backgroundColor: colors.accentAmber }]} />
        <Text style={[s.cardTitle, { color: colors.accentAmber }]}>TEMPERATURE</Text>
        {loading && <ActivityIndicator size="small" color={colors.accentAmber} style={{ marginLeft: 'auto' }} />}
      </View>

      {error ? (
        <Text style={s.errorText}>capteur indisponible</Text>
      ) : data ? (
        <View style={{ gap: 4 }}>
          <View style={{ flexDirection: 'row', alignItems: 'flex-end', gap: 4 }}>
            <Text style={s.bigTemp}>{data.temperature.toFixed(1)}</Text>
            <Text style={[s.barValue, { color: colors.accentAmber, marginBottom: 6 }]}>°C</Text>
          </View>
          <View style={{ flexDirection: 'row', gap: 24, marginTop: 2 }}>
            <View>
              <Text style={s.subLabel}>humidité</Text>
              <Text style={[s.barValue, { color: colors.accentHumidity }]}>{data.humidity.toFixed(0)}%</Text>
            </View>
            <View>
              <Text style={s.subLabel}>batterie</Text>
              <Text style={[s.barValue, { color: colors.accentBattery }]}>{data.battery}%</Text>
            </View>
          </View>
          <SparklineFallback points={history} />
          <Text style={[s.metaText, { marginTop: 2 }]}>{fmtTime(data.created_at)}</Text>
        </View>
      ) : null}
    </View>
  );
}

// ---- Prise Card ----

function CardPrise() {
  const [data, setData] = useState<PriseData | null>(null);
  const [toggling, setToggling] = useState(false);
  const { s, colors } = useTheme();

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
    <View style={[s.card, { backgroundColor: colors.cardEmer, borderColor: colors.borderEmer }]}>
      <View style={s.cardHeader}>
        <View style={[s.dot, { backgroundColor: colors.accentEmer }]} />
        <Text style={[s.cardTitle, { color: colors.accentEmer }]}>PRISE</Text>
      </View>
      <View style={{ gap: 4 }}>
        <View style={{ alignItems: 'center', marginVertical: 8 }}>
          <View style={[s.circle, on ? s.glowOn : s.glowOff]}>
            <Text style={[s.circleText, { color: on ? colors.stateOn : colors.stateOff }]}>{state}</Text>
          </View>
          <View style={[s.circleDot, { backgroundColor: on ? colors.stateOn : colors.stateOff }]} />
        </View>
        {data && <Text style={[s.metaText, { textAlign: 'center' }]}>{fmtTime(data.created_at)}</Text>}
        <Pressable
          onPress={toggle}
          disabled={toggling}
          style={({ pressed }) => [s.toggleBtn, pressed && { opacity: 0.7 }]}
        >
          {toggling ? (
            <ActivityIndicator size="small" color={colors.text} />
          ) : (
            <Text style={s.toggleBtnText}>Toggle</Text>
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
  const { s, colors } = useTheme();

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
  const color = contact == null ? '#64748b' : contact ? colors.contactClosed : colors.contactOpen;
  const dotColor = contact == null ? '#64748b' : contact ? colors.contactClosed : colors.contactOpen;
  const glowStyle = contact == null ? {} : contact ? s.glowOff : s.glowAlert;

  return (
    <View style={[s.card, { backgroundColor: colors.cardViolet, borderColor: colors.borderViolet }]}>
      <View style={s.cardHeader}>
        <View style={[s.dot, { backgroundColor: colors.accentViolet }]} />
        <Text style={[s.cardTitle, { color: colors.accentViolet }]}>PORTE</Text>
        {loading && <ActivityIndicator size="small" color={colors.accentViolet} style={{ marginLeft: 'auto' }} />}
      </View>
      <View style={{ gap: 4 }}>
        <View style={{ alignItems: 'center', marginVertical: 8 }}>
          <View style={[s.circle, glowStyle]}>
            <Text style={[s.circleText, { color, fontSize: 14 }]}>{label}</Text>
          </View>
          <View style={[s.circleDot, { backgroundColor: dotColor }]} />
        </View>
        {data?.battery != null && (
          <View style={{ marginTop: 4, alignItems: 'center' }}>
            <Text style={s.subLabel}>batterie</Text>
            <Text style={[s.barValue, { color: colors.accentBattery }]}>{data.battery}%</Text>
          </View>
        )}
        {data && <Text style={[s.metaText, { textAlign: 'center', marginTop: 2 }]}>{fmtTime(data.created_at)}</Text>}
      </View>
    </View>
  );
}

// ---- Root ----

export default function App() {
  return (
    <ThemeProvider>
      <AppInner />
    </ThemeProvider>
  );
}

function AppInner() {
  const [refreshing, setRefreshing] = useState(false);
  const [key, setKey] = useState(0);
  const { s, dark } = useTheme();

  const onRefresh = useCallback(() => {
    setRefreshing(true);
    setKey((k) => k + 1);
    setTimeout(() => setRefreshing(false), 600);
  }, []);

  return (
    <SafeAreaView style={s.safeArea}>
      <StatusBar style={dark ? 'light' : 'dark'} />
      <ScrollView
        contentContainerStyle={s.container}
        refreshControl={
          <RefreshControl refreshing={refreshing} onRefresh={onRefresh} tintColor="#2563eb" colors={['#2563eb']} />
        }
      >
        <View style={s.header}>
          <View>
            <View style={{ flexDirection: 'row', alignItems: 'baseline' }}>
              <Text style={s.logoNovo}>NOVO</Text>
              <Text style={[s.logoNovo, { color: '#2563eb' }]}>CEO</Text>
            </View>
            <View style={{ flexDirection: 'row', alignItems: 'center', gap: 6, marginTop: 3 }}>
              <View style={s.logoDash} />
              <Text style={s.logoSub}>GROUPE</Text>
              <View style={s.logoDash} />
            </View>
          </View>
          <View style={s.headerRight}>
            <ThemeSelector />
            <View style={s.statusDot} />
          </View>
        </View>

        <View key={key} style={s.grid}>
          <CardHealth />
          <CardTemperature />
          <CardPrise />
          <CardContact />
        </View>

        <Text style={s.footer}>novoceo · zigbee2mqtt monitoring</Text>
      </ScrollView>
    </SafeAreaView>
  );
}
