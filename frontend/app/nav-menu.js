import { LitElement, html, css } from 'https://esm.sh/lit';

class NavMenu extends LitElement {
  static styles = css`
    :host {
      display: block;
    }

    nav {
      background: var(--surface, #121212);
      border-bottom: 1px solid var(--hairline, #2a2a2a);
      display: flex;
      gap: var(--space-24, 24px);
      padding: var(--space-16, 16px) var(--space-24, 24px);
    }

    a {
      color: var(--fg, #eaeaea);
      font-family: var(--font-body, 'JetBrains Mono', monospace);
      font-size: var(--text-13, 13px);
      letter-spacing: 0.05em;
      text-decoration: none;
      text-transform: uppercase;
    }

    a:hover,
    a:focus {
      color: #ffffff;
    }

    a[aria-current] {
      color: #ffffff;
      border-bottom: 2px solid var(--accent, #ff2a2a);
    }

    a:focus-visible {
      outline: 2px solid var(--accent, #ff2a2a);
      outline-offset: 2px;
    }
  `;

  render() {
    const current = window.location.pathname;
    return html`
      <nav aria-label="Principal">
        <a href="/" ?aria-current="${current === '/' || current === '/index.html'}">Teste de Carga</a>
        <a href="/saldo.html" ?aria-current="${current === '/saldo.html'}">Saldo em Tempo Real</a>
      </nav>
    `;
  }
}

customElements.define('nav-menu', NavMenu);
