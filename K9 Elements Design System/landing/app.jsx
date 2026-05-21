// K9 Elements landing — app shell with Tweaks

const DEFAULTS = /*EDITMODE-BEGIN*/{
  "density": "comfortable",
  "heroAngle": -8,
  "heroOffset": -10,
  "discLayout": "table"
}/*EDITMODE-END*/;

function App() {
  const [t, setTweak] = useTweaks(DEFAULTS);

  React.useEffect(() => {
    document.documentElement.setAttribute("data-density", t.density);
    document.documentElement.setAttribute("data-disc-layout", t.discLayout);
  }, [t.density, t.discLayout]);

  return (
    <>
      <Utility />
      <Mast />
      <main className="k9l-shell">
        <Hero tweaks={t} />
        <Disciplines layout={t.discLayout} />
        <Flow />
        <Ticker />
        <Voice />
        <FinalCTA />
      </main>
      <Foot />

      <TweaksPanel title="Tweaks" subtitle="Landing page knobs">
        <TweakSection title="Density">
          <TweakRadio
            value={t.density}
            options={[{ label: "Compact", value: "compact" }, { label: "Comfortable", value: "comfortable" }, { label: "Editorial", value: "editorial" }]}
            onChange={(v) => setTweak("density", v)}
          />
        </TweakSection>
        <TweakSection title="Hero — slash placement">
          <TweakSlider label="Angle" value={t.heroAngle} min={-30} max={30} step={1} onChange={(v) => setTweak("heroAngle", v)} suffix="°" />
          <TweakSlider label="Offset" value={t.heroOffset} min={-40} max={40} step={1} onChange={(v) => setTweak("heroOffset", v)} suffix="%" />
        </TweakSection>
        <TweakSection title="Disciplines layout">
          <TweakRadio
            value={t.discLayout}
            options={[{ label: "Table", value: "table" }, { label: "Cards", value: "cards" }, { label: "Index", value: "index" }]}
            onChange={(v) => setTweak("discLayout", v)}
          />
        </TweakSection>
      </TweaksPanel>
    </>
  );
}

ReactDOM.createRoot(document.getElementById("root")).render(<App />);
