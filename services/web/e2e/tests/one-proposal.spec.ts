import { expect, test, type APIRequestContext, type Page } from '@playwright/test';

const apiURL = process.env.API_URL ?? 'http://localhost:8080';
const screenshotsDir = process.env.SCREENSHOTS_DIR ?? '../../../reports/screenshots';

// A console error is a failure, except the ones listed as expected.
function failOnConsoleErrors(page: Page, expected: RegExp[] = []): string[] {
  const errors: string[] = [];
  page.on('console', (msg) => {
    if (msg.type() === 'error' && !expected.some((re) => re.test(msg.text()))) errors.push(msg.text());
  });
  page.on('pageerror', (err) => errors.push(err.message));
  return errors;
}

const day = (offset: number) => new Date(Date.now() + offset * 86_400_000).toISOString().slice(0, 10);

// The organizer side is set up through the api: the actor under test is the participant.
async function eventInParticipation(request: APIRequestContext): Promise<string> {
  const signUp = await request.post(`${apiURL}/api/v1/users`, {
    data: { name: 'Organizadora', email: `organizer.${Date.now()}@example.com`, password: 'e2e-password-123' },
  });
  expect(signUp.status()).toBe(201);
  const headers = { Authorization: `Bearer ${(await signUp.json()).token as string}` };
  const created = await request.post(`${apiURL}/api/v1/events`, {
    headers,
    data: {
      name: `Propuesta única ${Date.now()}`,
      description: 'Evento del test end to end de una propuesta por inscrito.',
      start_date: day(1),
      end_date: day(2),
      max_participants: 10,
    },
  });
  expect(created.status()).toBe(201);
  const eventId = (await created.json()).event.id as string;
  const opened = await request.patch(`${apiURL}/api/v1/events/${eventId}/stage`, {
    headers,
    data: { stage: 'participation', estimated_end_date: day(7) },
  });
  expect(opened.status()).toBe(200);
  return eventId;
}

const proposal = (name: string) => ({ name, mimeType: 'text/plain', buffer: Buffer.from(`Contenido de ${name}`) });

const uploadResponse = (page: Page) =>
  page.waitForResponse((r) => r.url().includes('/attachment') && r.request().method() === 'POST');

// A participant with the event open in two tabs uploads from both: the second upload is refused
// with the INS-E15 message, and the event keeps only the first proposal.
test('[INS-05][INS-E15][INV-13] a participant who already uploaded a proposal is told so when uploading again', async ({ page, context, request }) => {
  const eventId = await eventInParticipation(request);
  const errors = failOnConsoleErrors(page);

  await page.goto('/register');
  await page.getByLabel('Nombre completo').fill('Participante E2E');
  await page.getByLabel('Email').fill(`participant.${Date.now()}@example.com`);
  await page.getByPlaceholder('Al menos 8 caracteres').fill('e2e-password-123');
  await page.getByRole('button', { name: 'Crear cuenta' }).click();
  await expect(page).toHaveURL(/\/events$/);

  await page.goto(`/events/${eventId}`);
  await page.getByRole('button', { name: 'Inscribirme al evento' }).click();
  await expect(page.getByRole('heading', { level: 2, name: 'Sube tu propuesta' })).toBeVisible();

  // The same participant in a second tab, before the first upload.
  const second = await context.newPage();
  // The browser logs the refused request, and the web's HTTP client and upload service log the api
  // error they handle.
  const secondErrors = failOnConsoleErrors(second, [/status of 409/, /Participant already has an attachment for this event/]);
  await second.goto(`/events/${eventId}`);
  await expect(second.getByRole('heading', { level: 2, name: 'Sube tu propuesta' })).toBeVisible();

  await page.getByLabel(/Arrastra tu archivo/).setInputFiles(proposal('primera.txt'));
  const accepted = uploadResponse(page);
  await page.getByRole('button', { name: 'Enviar propuesta' }).click();
  expect((await accepted).status()).toBe(201);

  await second.getByLabel(/Arrastra tu archivo/).setInputFiles(proposal('segunda.txt'));
  const refused = uploadResponse(second);
  await second.getByRole('button', { name: 'Enviar propuesta' }).click();
  expect((await refused).status()).toBe(409);
  await expect(second.getByRole('alert')).toHaveText('Ya subiste una propuesta a este evento.');

  await second.setViewportSize({ width: 1280, height: 900 });
  await second.screenshot({ path: `${screenshotsDir}/662/event-detail-duplicate-desktop.png`, fullPage: true });
  await second.setViewportSize({ width: 390, height: 844 });
  await expect(second.getByRole('alert')).toHaveText('Ya subiste una propuesta a este evento.');
  await second.screenshot({ path: `${screenshotsDir}/662/event-detail-duplicate-phone.png`, fullPage: true });

  // The participant has only the first proposal in the event.
  const token = await page.evaluate(() => localStorage.getItem('openparallax_token'));
  expect(token).toBeTruthy();
  const list = await request.get(`${apiURL}/api/v1/events/${eventId}/attachments`, {
    headers: { Authorization: `Bearer ${token}` },
  });
  expect(list.status()).toBe(200);
  const names = ((await list.json()).data as Array<{ original_name: string }>).map((a) => a.original_name);
  expect(names).toEqual(['primera.txt']);

  await second.close();
  expect(errors, 'console errors of the first tab').toEqual([]);
  expect(secondErrors, 'console errors of the second tab').toEqual([]);
});
