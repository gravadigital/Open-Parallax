import { expect, test, type Page } from '@playwright/test';

// A console error while rendering is a failure, not a warning.
function failOnConsoleErrors(page: Page): string[] {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error') errors.push(msg.text());
  });
  page.on('pageerror', (err) => errors.push(err.message));
  return errors;
}

const unique = (prefix: string) => `${prefix} ${Date.now()}`;

// The main path of the web surface: a person creates an account, creates an event
// with the wizard, opens registration, and anyone (without a session) can see the
// event and the button to register.
test('an organizer creates an event, opens registration and the public sees it', async ({ page, browser }) => {
  const errors = failOnConsoleErrors(page);
  const email = `organizer.${Date.now()}@example.com`;

  await page.goto('/register');
  await page.getByLabel('Nombre completo').fill('Organizadora E2E');
  await page.getByLabel('Email').fill(email);
  await page.getByPlaceholder('Al menos 8 caracteres').fill('e2e-password-123');
  await page.getByRole('button', { name: 'Crear cuenta' }).click();
  await expect(page).toHaveURL(/\/events$/);

  const name = unique('Concurso e2e');
  await page.goto('/events/create');
  await page.getByLabel('Nombre del evento').fill(name);
  await page.getByLabel('Descripción').fill('Evento creado por el test end to end de la web.');
  await page.getByRole('button', { name: 'Siguiente: Fechas y cupo →' }).click();
  await page.getByRole('button', { name: 'Siguiente: Revisar →' }).click();
  await page.getByRole('button', { name: 'Crear evento', exact: true }).click();

  await expect(page).toHaveURL(/\/events\/[0-9a-f-]+\/manage$/);
  await expect(page.getByText('Evento creado. Cuando esté listo, abre la inscripción.')).toBeVisible();
  const eventPath = new URL(page.url()).pathname.replace(/\/manage$/, '');

  await page.getByRole('button', { name: 'Abrir inscripción →' }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Abrir inscripción', exact: true }).click();
  await expect(page.getByText(/^Inscripción abierta/).first()).toBeVisible();

  const visitor = await browser.newPage();
  const visitorErrors = failOnConsoleErrors(visitor);
  await visitor.goto(eventPath);
  await expect(visitor.getByRole('heading', { name })).toBeVisible();
  await expect(visitor.getByRole('button', { name: 'Inscribirme al evento' })).toBeVisible();
  await visitor.close();

  expect(errors, 'console errors of the organizer').toEqual([]);
  expect(visitorErrors, 'console errors of the visitor').toEqual([]);
});
