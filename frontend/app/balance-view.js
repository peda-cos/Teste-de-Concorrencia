import { LitElement, html, css } from 'https://esm.sh/lit';

class BalanceView extends LitElement {
  static properties = {
    balance: { type: Number },
    pending: { type: Boolean },
  };

  constructor() {
    super();
    this.balance = 0;
    this.pending = false;
    this.socket = null;
  }

  static styles = css`
    :host {
      display: block;
    }

    .readout {
      border: 1px solid var(--hairline, #2a2a2a);
      margin-bottom: var(--space-24, 24px);
      padding: var(--space-32, 32px) var(--space-24, 24px);
      text-align: center;
    }

    .label {
      display: block;
      font-size: var(--text-12, 12px);
      margin-bottom: var(--space-16, 16px);
      text-transform: uppercase;
    }

    .balance {
      color: var(--terminal, #4af626);
      font-family: var(--font-display, 'Inter', sans-serif);
      font-size: var(--text-display, clamp(4rem, 10vw, 15rem));
      font-weight: 900;
      letter-spacing: -0.05em;
      line-height: 0.9;
      transition: none;
    }

    .actions {
      display: flex;
      gap: var(--space-16, 16px);
    }
  `;

  connectedCallback() {
    super.connectedCallback();
    this.fetchBalance();
    this.connect();
  }

  disconnectedCallback() {
    super.disconnectedCallback();
    if (this.socket) {
      this.socket.close();
      this.socket = null;
    }
  }

  connect() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const url = `${protocol}//${window.location.host}/ws`;
    this.socket = new WebSocket(url);

    this.socket.addEventListener('message', (event) => {
      try {
        const data = JSON.parse(event.data);
        if (Number.isFinite(data.balance)) {
          this.balance = data.balance;
        }
      } catch (err) {
        // ignore malformed message
      }
    });

    this.socket.addEventListener('error', () => {
      // closed connection is handled by close listener
    });

    this.socket.addEventListener('close', () => {
      this.socket = null;
    });
  }

  async fetchBalance() {
    try {
      const response = await fetch('/balance');
      if (!response.ok) {
        return;
      }
      const data = await response.json();
      if (Number.isFinite(data.balance)) {
        this.balance = data.balance;
      }
    } catch (err) {
      // ignore initial fetch failure; websocket will catch up
    }
  }

  async postOperation(endpoint) {
    if (this.pending) {
      return;
    }
    this.pending = true;
    try {
      const response = await fetch(endpoint, { method: 'POST' });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      const data = await response.json();
      if (Number.isFinite(data.balance)) {
        this.balance = data.balance;
      }
    } catch (err) {
      // keep current balance; websocket will resync
    } finally {
      this.pending = false;
    }
  }

  render() {
    return html`
      <div class="readout">
        <span class="label">Saldo atual</span>
        <span class="balance" aria-live="polite">${this.balance}</span>
      </div>
      <div class="actions">
        <button type="button" ?disabled="${this.pending}" @click="${() => this.postOperation('/credit')}">Creditar +1</button>
        <button type="button" ?disabled="${this.pending}" @click="${() => this.postOperation('/debit')}">Debitar -1</button>
      </div>
    `;
  }
}

customElements.define('balance-view', BalanceView);
