import { LitElement, html, css } from 'https://esm.sh/lit';
import { buttonVariants } from './button-variants.js';

class LoadTestConfig extends LitElement {
  static properties = {
    groups: { type: Array },
    loading: { type: Boolean },
    metrics: { type: Object },
  };

  constructor() {
    super();
    this.groups = [];
    this.loading = false;
    this.metrics = null;
  }

  static styles = [css`
    :host {
      display: block;
    }

    .row {
      align-items: end;
      border: 1px solid var(--hairline, #2a2a2a);
      display: grid;
      gap: var(--space-16, 16px);
      grid-template-columns: 1fr 1fr auto auto;
      margin-bottom: var(--space-16, 16px);
      padding: var(--space-16, 16px);
    }

    .field {
      display: flex;
      flex-direction: column;
    }

    .actions {
      display: flex;
      gap: var(--space-16, 16px);
      margin-top: var(--space-24, 24px);
    }

    .metrics {
      display: grid;
      gap: var(--space-16, 16px);
      grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
      margin-top: var(--space-32, 32px);
    }

    .metric {
      border: 1px solid var(--hairline, #2a2a2a);
      padding: var(--space-16, 16px);
    }

    .metric-value {
      color: var(--terminal, #4af626);
      display: block;
      font-size: var(--text-32, 32px);
      margin-top: var(--space-8, 8px);
    }

    .empty {
      border: 1px dashed var(--hairline, #2a2a2a);
      padding: var(--space-24, 24px);
    }
  `, buttonVariants];

  addGroup() {
    this.groups = [...this.groups, { users: 1, requests: 10, type: 'credit' }];
  }

  removeGroup(index) {
    const next = [...this.groups];
    next.splice(index, 1);
    this.groups = next;
  }

  updateGroup(index, field, value) {
    const next = [...this.groups];
    next[index] = { ...next[index], [field]: value };
    this.groups = next;
  }

  async startTest() {
    if (this.groups.length === 0) {
      return;
    }
    for (let i = 0; i < this.groups.length; i++) {
      const g = this.groups[i];
      if (!g.users || g.users < 1) {
        this.metrics = { error: `Grupo ${i + 1}: número de usuários deve ser maior que zero` };
        return;
      }
      if (!g.requests || g.requests < 1) {
        this.metrics = { error: `Grupo ${i + 1}: número de requisições deve ser maior que zero` };
        return;
      }
    }
    this.loading = true;
    this.metrics = null;
    try {
      const response = await fetch('/load-test/start', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify(this.groups),
      });
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }
      this.metrics = await response.json();
    } catch (err) {
      this.metrics = { error: err.message };
    } finally {
      this.loading = false;
    }
  }

  render() {
    return html`
      ${this.groups.length === 0
        ? html`<div class="empty" role="status">Nenhum grupo configurado. Adicione um grupo para iniciar.</div>`
        : this.groups.map((group, index) => html`
            <div class="row">
              <div class="field">
                <label for="users-${index}">Usuários</label>
                <input
                  id="users-${index}"
                  type="number"
                  min="1"
                  step="1"
                  .value="${String(group.users)}"
                  @input="${(e) => this.updateGroup(index, 'users', Number(e.target.value))}"
                />
              </div>
              <div class="field">
                <label for="requests-${index}">Requisições / usuário</label>
                <input
                  id="requests-${index}"
                  type="number"
                  min="1"
                  step="1"
                  .value="${String(group.requests)}"
                  @input="${(e) => this.updateGroup(index, 'requests', Number(e.target.value))}"
                />
              </div>
              <div class="field">
                <label for="type-${index}">Operação</label>
                <select
                  id="type-${index}"
                  .value="${group.type}"
                  @change="${(e) => this.updateGroup(index, 'type', e.target.value)}"
                >
                  <option value="credit">Crédito</option>
                  <option value="debit">Débito</option>
                </select>
              </div>
              <button type="button" class="btn-danger" @click="${() => this.removeGroup(index)}" ?disabled="${this.loading}">Remover</button>
            </div>
          `)}

      <div class="actions">
        <button type="button" class="btn-secondary" @click="${this.addGroup}" ?disabled="${this.loading}">Adicionar Grupo</button>
        <button type="button" class="btn-primary" @click="${this.startTest}" ?disabled="${this.loading || this.groups.length === 0}">${this.loading ? 'Executando…' : 'Iniciar Teste'}</button>
      </div>

      ${this.metrics && !this.metrics.error
        ? html`
            <section class="metrics" aria-label="Resultados">
              <div class="metric">
                <span>Total de requisições</span>
                <span class="metric-value">${this.metrics.totalRequests}</span>
              </div>
              <div class="metric">
                <span>Sucessos</span>
                <span class="metric-value">${this.metrics.successes}</span>
              </div>
              <div class="metric">
                <span>Falhas</span>
                <span class="metric-value">${this.metrics.failures}</span>
              </div>
              <div class="metric">
                <span>Duração (ms)</span>
                <span class="metric-value">${this.metrics.durationMs}</span>
              </div>
              <div class="metric">
                <span>Saldo final</span>
                <span class="metric-value">${this.metrics.finalBalance}</span>
              </div>
            </section>
          `
        : ''}
      ${this.metrics && this.metrics.error
        ? html`<div class="empty" role="alert">Erro: ${this.metrics.error}</div>`
        : ''}
    `;
  }
}

customElements.define('load-test-config', LoadTestConfig);
