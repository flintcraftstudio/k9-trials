/* eslint-disable */
// App — top-level page router. Holds the current route in state and renders
// the matching page. window.kit.navigate(route) is the public API used by
// links throughout the kit.

function App() {
  const [route, setRoute] = React.useState('home');
  const [openTrial, setOpenTrial] = React.useState(null);

  React.useEffect(() => {
    window.kit = {
      navigate: (r) => { setRoute(r); setOpenTrial(null); window.scrollTo({ top: 0, behavior: 'instant' }); },
      openTrial: (t) => setOpenTrial(t),
    };
  }, []);

  const page = (() => {
    switch (route) {
      case 'home':
        return (
          <main className="fade-up">
            <Hero />
            <EventGrid />
            <TrialListSection eyebrow="This season" title="Upcoming trials" limit={3} />
            <Testimonials />
            <PricingTiers />
          </main>
        );
      case 'trials':
        return (
          <main className="fade-up">
            <section className="section-lg" data-screen-label="03 Trials Index">
              <Container>
                <div className="flex-col gap-4" style={{ maxWidth: 720 }}>
                  <Eyebrow>Trial schedule</Eyebrow>
                  <h1 className="hero-h1">Every K9 Elements trial, in one schedule.</h1>
                </div>
              </Container>
            </section>
            <TrialListSection title="Spring & Summer 2026" eyebrow="" />
          </main>
        );
      case 'training':
        return <TrainingPage />;
      case 'pricing':
        return (
          <main className="fade-up">
            <PricingTiers />
            <section className="section" data-screen-label="07 CTA">
              <Container>
                <div className="flex-col gap-6" style={{ maxWidth: 640 }}>
                  <Eyebrow>Questions</Eyebrow>
                  <h2 className="h2">Talk to someone who runs trials.</h2>
                  <p className="lede">We’ll answer questions about hosting, judging, and how the platform handles entries.</p>
                  <div className="flex gap-3">
                    <Button variant="primary" size="lg">Email us</Button>
                    <Button variant="plain" size="lg">Book a call <Icon name="arrow-right" /></Button>
                  </div>
                </div>
              </Container>
            </section>
          </main>
        );
      default:
        return null;
    }
  })();

  return (
    <React.Fragment>
      <NavBar route={route} />
      {page}
      <Footer />
      <TrialDetailDrawer trial={openTrial} onClose={() => setOpenTrial(null)} />
    </React.Fragment>
  );
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);
