import { css } from 'https://esm.sh/lit';

export const buttonVariants = css`
  button {
    border-radius: 0;
    border: 1px solid var(--hairline, #2a2a2a);
    box-sizing: border-box;
    cursor: pointer;
    font-family: inherit;
    font-size: inherit;
    letter-spacing: inherit;
    min-height: 44px;
    padding: var(--space-12, 12px) var(--space-16, 16px);
    text-transform: uppercase;
    transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }

  .btn-primary {
    background: var(--accent, #ff2a2a);
    border-color: var(--accent, #ff2a2a);
    color: var(--bg, #0a0a0a);
  }

  .btn-primary:hover:not(:disabled) {
    color: #ffffff;
    border-color: #ffffff;
  }

  .btn-primary:active:not(:disabled) {
    background: var(--bg, #0a0a0a);
    border-color: var(--accent, #ff2a2a);
    color: var(--accent, #ff2a2a);
  }

  .btn-secondary {
    background: var(--surface, #121212);
    border-color: var(--hairline, #2a2a2a);
    color: var(--fg, #eaeaea);
  }

  .btn-secondary:hover:not(:disabled) {
    color: #ffffff;
    border-color: var(--fg, #eaeaea);
  }

  .btn-secondary:active:not(:disabled) {
    background: var(--accent, #ff2a2a);
    border-color: var(--accent, #ff2a2a);
    color: var(--bg, #0a0a0a);
  }

  .btn-danger {
    background: var(--surface, #121212);
    border-color: var(--accent, #ff2a2a);
    color: var(--accent, #ff2a2a);
  }

  .btn-danger:hover:not(:disabled) {
    color: #ffffff;
    border-color: #ffffff;
  }

  .btn-danger:active:not(:disabled) {
    background: var(--accent, #ff2a2a);
    border-color: var(--accent, #ff2a2a);
    color: var(--bg, #0a0a0a);
  }

  .btn-success {
    background: var(--terminal, #4af626);
    border-color: var(--terminal, #4af626);
    color: var(--bg, #0a0a0a);
  }

  .btn-success:hover:not(:disabled) {
    color: #ffffff;
    border-color: #ffffff;
  }

  .btn-success:active:not(:disabled) {
    background: var(--bg, #0a0a0a);
    border-color: var(--terminal, #4af626);
    color: var(--terminal, #4af626);
  }

  .btn-ghost {
    background: transparent;
    border-color: transparent;
    color: var(--fg, #eaeaea);
  }

  .btn-ghost:hover:not(:disabled) {
    color: #ffffff;
    border-color: var(--hairline, #2a2a2a);
  }

  .btn-ghost:active:not(:disabled) {
    background: var(--accent, #ff2a2a);
    border-color: var(--accent, #ff2a2a);
    color: var(--bg, #0a0a0a);
  }

  .btn-primary:disabled,
  .btn-secondary:disabled,
  .btn-danger:disabled,
  .btn-success:disabled,
  .btn-ghost:disabled {
    background: var(--surface, #121212);
    border-color: var(--hairline, #2a2a2a);
    color: #555555;
    cursor: not-allowed;
  }

  .btn-ghost:disabled {
    background: transparent;
  }

  .btn-primary:focus-visible,
  .btn-secondary:focus-visible,
  .btn-danger:focus-visible,
  .btn-success:focus-visible,
  .btn-ghost:focus-visible {
    outline: 2px solid var(--accent, #ff2a2a);
    outline-offset: 2px;
  }
`;
