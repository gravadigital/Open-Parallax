import React from 'react';
import fs from 'fs';
import path from 'path';
import { screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import AppFooter from './AppFooter';
import { renderWithProviders } from '../../../test-utils/renderWithProviders';

describe('AppFooter', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.lang = '';
  });

  it('TS-1: marca, versión, repositorio y contacto en español', () => {
    renderWithProviders(<AppFooter />, { locale: 'es' });

    const footer = screen.getByRole('contentinfo');
    expect(within(footer).getByRole('link', { name: 'TELESCOPIO' })).toHaveAttribute('href', '/');
    expect(within(footer).getByText('V1.0')).toBeInTheDocument();
    const repo = within(footer).getByRole('link', { name: /gravadigital\/telescope/ });
    expect(repo).toHaveAttribute('href', 'https://github.com/gravadigital/telescope');
    expect(repo).toHaveAttribute('target', '_blank');
    expect(repo).toHaveAttribute('rel', 'noopener noreferrer');
    expect(within(footer).getByText('CONTACTO')).toBeInTheDocument();
    expect(within(footer).getByRole('link', { name: 'info@grava.digital' })).toHaveAttribute(
      'href',
      'mailto:info@grava.digital'
    );
    expect(within(footer).getByRole('link', { name: '+54 9 11 3094 2020' })).toHaveAttribute(
      'href',
      'tel:+5491130942020'
    );
  });

  it('TS-2: textos en inglés', () => {
    renderWithProviders(<AppFooter />, { locale: 'en' });

    const footer = screen.getByRole('contentinfo');
    expect(within(footer).getByRole('link', { name: 'TELESCOPIO' })).toHaveAttribute('href', '/');
    expect(within(footer).getByText('CONTACT')).toBeInTheDocument();
  });

  it('TS-3: la marca navega con el router, sin recargar', async () => {
    renderWithProviders(<AppFooter />, { locale: 'es', route: '/events' });
    expect(screen.getByTestId('location')).toHaveTextContent('/events');

    await userEvent.click(screen.getByRole('link', { name: 'TELESCOPIO' }));
    expect(screen.getByTestId('location')).toHaveTextContent('/');
  });
});

describe('AppFooter.css', () => {
  const read = (file: string) =>
    fs.readFileSync(path.join(__dirname, file), 'utf8').replace(/\/\*[\s\S]*?\*\//g, '');
  const css = read('AppFooter.css');
  const headerCss = read('../app-header/AppHeader.css');

  const mediaStart = css.indexOf('@media');
  const base = css.slice(0, mediaStart);
  const media = css.slice(mediaStart);
  const rule = (source: string, selector: string) =>
    new RegExp(`${selector.replace(/\./g, '\\.')}\\s*\\{([^}]*)\\}`).exec(source)?.[1] ?? '';
  const declaration = (block: string, property: string) =>
    new RegExp(`${property}:\\s*([^;]+);`).exec(block)?.[1].trim();

  it('TS-4: mismo fondo que el header', () => {
    const footerBg = declaration(rule(css, '.ly-footer'), 'background');
    const headerBg = declaration(rule(headerCss, '.ly-header'), 'background');

    expect(footerBg).toBe('var(--bg-band)');
    expect(footerBg).toBe(headerBg);
    expect(declaration(rule(css, '.ly-footer'), 'color')).toBe('var(--text-inverse)');
  });

  it('TS-5: foco visible con el token de foco', () => {
    expect(css).toMatch(/\.ly-footer__logo:focus-visible/);
    expect(css).toMatch(/\.ly-footer__github:focus-visible/);
    expect(css).toMatch(/\.ly-footer__contact a:focus-visible/);
    expect(css).toMatch(/box-shadow:\s*var\(--focus-ring\)/);
  });

  it('TS-6: mobile primero, en columna, y grilla desde 768px', () => {
    expect(declaration(rule(base, '.ly-footer__inner'), 'flex-direction')).toBe('column');
    expect(media).toMatch(/@media \(min-width: 768px\)/);
    expect(declaration(rule(media, '.ly-footer__inner'), 'display')).toBe('grid');
  });
});
