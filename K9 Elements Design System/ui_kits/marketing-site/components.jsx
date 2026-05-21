/* eslint-disable */
// Primitives — buttons, pills, eyebrow, container, navbar, footer, screenshot frame.

function Container({ children, className = '' }) {
  return <div className={`container ${className}`}>{children}</div>;
}

function Button({ children, variant = 'primary', size = 'md', href, onClick, className = '' }) {
  const cls = `btn ${size} btn-${variant} ${className}`;
  if (href) return <a href={href} className={cls} onClick={onClick}>{children}</a>;
  return <button type="button" className={cls} onClick={onClick}>{children}</button>;
}

function Pill({ children, variant = 'open' }) {
  return <span className={`pill pill-${variant}`}>
    {variant !== 'closed' && variant !== 'info' && variant.startsWith('event-') ? null : <span className="dot" />}
    {children}
  </span>;
}

function EventPill({ event }) {
  const label = event[0].toUpperCase() + event.slice(1);
  return <span className={`pill pill-event-${event}`}>{label}</span>;
}

function Eyebrow({ children }) { return <div className="eyebrow">{children}</div>; }

function Icon({ name, size = 13, className = '', style }) {
  const isSocial = name.startsWith('social-');
  const px = size;
  return <img
    src={`../../assets/icons/${name}.svg`}
    alt=""
    aria-hidden="true"
    className={className}
    style={{ width: px, height: px, ...style }}
  />;
}

function Logo({ size = 32 }) {
  return <a href="#home" className="logo" onClick={(e) => { e.preventDefault(); window.kit.navigate('home'); }}>
    <img src="../../assets/k9-elements-logo.png" alt="" style={{ width: size, height: size }} />
    <span>K9 Elements</span>
  </a>;
}

function NavBar({ route }) {
  const links = [
    { id: 'home',     label: 'Home' },
    { id: 'trials',   label: 'Trials' },
    { id: 'training', label: 'Training' },
    { id: 'pricing',  label: 'Pricing' },
  ];
  return (
    <header className="navbar">
      <div className="inner">
        <Logo />
        <nav className="links" style={{ flex: 1, justifyContent: 'center' }}>
          {links.map(l => (
            <a
              key={l.id}
              href={`#${l.id}`}
              className={route === l.id ? 'active' : ''}
              onClick={(e) => { e.preventDefault(); window.kit.navigate(l.id); }}
            >{l.label}</a>
          ))}
        </nav>
        <div className="actions">
          <Button variant="plain" size="md">Log in</Button>
          <Button variant="primary" size="md">Get started</Button>
        </div>
      </div>
    </header>
  );
}

function Footer() {
  return (
    <footer className="footer">
      <div className="inner">
        <Container>
          <div className="footer-grid">
            <div className="flex-col gap-3">
              <Logo />
              <p className="body" style={{ maxWidth: 360, marginTop: 8 }}>
                Organizing trials and supplying training material for working-dog sport. Four events, one schedule.
              </p>
              <div className="flex gap-4" style={{ marginTop: 16 }}>
                <a href="#" aria-label="Instagram"><Icon name="social-instagram" size={20} style={{ color: 'var(--mist-700)' }} /></a>
                <a href="#" aria-label="YouTube"><Icon name="social-youtube" size={20} style={{ color: 'var(--mist-700)' }} /></a>
                <a href="#" aria-label="X"><Icon name="social-x" size={20} style={{ color: 'var(--mist-700)' }} /></a>
              </div>
            </div>
            <div className="footer-cats">
              <div>
                <h4>Sport</h4>
                <ul>
                  <li><a href="#">Trial schedule</a></li>
                  <li><a href="#">Find a club</a></li>
                  <li><a href="#">Judge directory</a></li>
                  <li><a href="#">Rulebooks</a></li>
                </ul>
              </div>
              <div>
                <h4>Training</h4>
                <ul>
                  <li><a href="#">Obedience path</a></li>
                  <li><a href="#">Tracking path</a></li>
                  <li><a href="#">Detection path</a></li>
                  <li><a href="#">Protection path</a></li>
                </ul>
              </div>
              <div>
                <h4>Members</h4>
                <ul>
                  <li><a href="#">Sign up</a></li>
                  <li><a href="#">Log in</a></li>
                  <li><a href="#">Pricing</a></li>
                  <li><a href="#">Help</a></li>
                </ul>
              </div>
              <div>
                <h4>K9 Elements</h4>
                <ul>
                  <li><a href="#">About</a></li>
                  <li><a href="#">Contact</a></li>
                  <li><a href="#">Press</a></li>
                  <li><a href="#">Careers</a></li>
                </ul>
              </div>
            </div>
          </div>
          <div className="fine">
            <div>© 2026 K9 Elements</div>
            <div>Reston, VA · sport@k9elements.com</div>
          </div>
        </Container>
      </div>
    </footer>
  );
}

function PhotoPlaceholder({ kind = 'field', caption, height = '60vw', maxHeight = 640, play = true, eventTint }) {
  const tintStyle = eventTint ? {
    boxShadow: `inset 0 0 0 9999px var(--${eventTint}-600)`,
    mixBlendMode: 'overlay',
    opacity: 0.15,
  } : null;
  return (
    <div className={`photo photo-${kind}`} style={{ height, maxHeight, width: '100%' }}>
      {eventTint && <div style={{ position: 'absolute', inset: 0, background: `var(--${eventTint}-600)`, mixBlendMode: 'overlay', opacity: 0.18 }} />}
      {play && (
        <div className="play"><div className="circle" /></div>
      )}
      {caption && <div className="caption">{caption}</div>}
    </div>
  );
}

Object.assign(window, {
  Container, Button, Pill, EventPill, Eyebrow, Icon, Logo, NavBar, Footer, PhotoPlaceholder,
});
