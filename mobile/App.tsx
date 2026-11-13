import { StatusBar } from "expo-status-bar";
import { useState } from "react";
import { Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { checkReadiness } from "./src/lib/api";

export default function App() {
  const [status, setStatus] = useState("Connection not checked");
  const [loading, setLoading] = useState(false);
  async function check() {
    setLoading(true);
    setStatus("Checking connection…");
    try {
      setStatus(await checkReadiness());
    } catch {
      setStatus("Unable to connect. Check your network and API address.");
    } finally {
      setLoading(false);
    }
  }
  return (
    <SafeAreaProvider>
      <SafeAreaView style={styles.safeArea}>
        <StatusBar style="dark" />
        <ScrollView contentContainerStyle={styles.content}>
          <View style={styles.brand}>
            <Text style={styles.mark}>C</Text>
            <Text style={styles.brandName}>CLIMS</Text>
          </View>
          <Text style={styles.eyebrow}>SUPPLY OPERATIONS · MOBILE</Text>
          <Text style={styles.title}>Closer to{`\n`}the field.</Text>
          <Text style={styles.description}>
            A connected foundation for the teams keeping supplies moving.
          </Text>
          <View style={styles.card}>
            <Text style={styles.label}>PLATFORM CONNECTION</Text>
            <Text accessibilityLiveRegion="polite" style={styles.status}>
              {status}
            </Text>
            <Pressable
              accessibilityRole="button"
              disabled={loading}
              onPress={check}
              style={[styles.button, loading && styles.disabled]}
            >
              <Text style={styles.buttonText}>{loading ? "Checking…" : "Check connection ↗"}</Text>
            </Pressable>
          </View>
          <View style={styles.note}>
            <Text style={styles.noteTitle}>A foundation for the work ahead.</Text>
            <Text style={styles.noteText}>
              The mobile project is ready to build on, with the CLIMS API behind one shared gateway.
            </Text>
          </View>
        </ScrollView>
      </SafeAreaView>
    </SafeAreaProvider>
  );
}

const styles = StyleSheet.create({
  safeArea: { flex: 1, backgroundColor: "#f3f6f4" },
  content: { padding: 28, paddingBottom: 48 },
  brand: { flexDirection: "row", alignItems: "center", gap: 12, marginBottom: 64 },
  mark: {
    backgroundColor: "#096d60",
    color: "white",
    fontSize: 26,
    fontWeight: "700",
    width: 42,
    height: 42,
    textAlign: "center",
    lineHeight: 42,
    borderRadius: 10,
  },
  brandName: { fontSize: 25, fontWeight: "800", color: "#142a34", letterSpacing: -1 },
  eyebrow: { color: "#096d60", fontSize: 10, fontWeight: "700", letterSpacing: 1.5 },
  title: {
    fontSize: 48,
    lineHeight: 52,
    letterSpacing: -2,
    color: "#142a34",
    marginTop: 22,
    fontWeight: "500",
  },
  description: { fontSize: 16, lineHeight: 26, color: "#64757c", marginTop: 20, marginBottom: 36 },
  card: {
    padding: 24,
    backgroundColor: "white",
    borderWidth: 1,
    borderColor: "#dce5e4",
    borderRadius: 12,
  },
  label: { fontSize: 10, letterSpacing: 1.4, color: "#64757c", fontWeight: "700" },
  status: { fontSize: 16, lineHeight: 24, color: "#142a34", marginTop: 14, marginBottom: 24 },
  button: { backgroundColor: "#142a34", padding: 16, borderRadius: 7, alignItems: "center" },
  buttonText: { color: "white", fontSize: 14, fontWeight: "600" },
  disabled: { opacity: 0.6 },
  note: { marginTop: 32, paddingTop: 24, borderTopWidth: 1, borderTopColor: "#dce5e4" },
  noteTitle: { fontSize: 14, fontWeight: "600", color: "#142a34", marginBottom: 8 },
  noteText: { fontSize: 13, lineHeight: 22, color: "#64757c" },
});
