/**
 * Single source of truth for legal-page metadata.
 *
 * The backend (Go) does NOT read this file: the frontend sends the
 * `version` fields in /auth/register (Spec 018 §7) so the Go side
 * doesn't need to know about doc versions.
 *
 * `path` points to the .md file in docs/legal/ relative to the repo
 * root. `src/lib/legal/docs.ts` imports these via Vite `?raw` at
 * build time so there is no mirrored `src/lib/legal/content/` to
 * keep in sync (Spec 018 decision #13).
 */

export const CONTACT = {
	general: 'hello@teren.dev',
	privacy: 'privacy@teren.dev',
} as const;

export const LEGAL_DOCS = {
	terms: {
		es: {
			version: '1.0',
			updated: '2026-08-31',
			title: 'Términos de Servicio de Itinera',
			description:
				'Condiciones de uso, propiedad intelectual, limitación de responsabilidad y jurisdicción aplicable al servicio Itinera.',
			path: '../../../../docs/legal/TERMS_OF_SERVICE.md',
		},
		en: {
			version: '1.0',
			updated: '2026-08-31',
			title: 'Terms of Service for Itinera',
			description:
				'Conditions of use, intellectual property, limitation of liability, and applicable jurisdiction for the Itinera service.',
			path: '../../../../docs/legal/TERMS_OF_SERVICE_EN.md',
		},
	},
	privacy: {
		es: {
			version: '1.0',
			updated: '2026-08-31',
			title: 'Política de Privacidad de Itinera',
			description:
				'Información sobre el tratamiento de datos personales en Itinera conforme al RGPD y la LOPDGDD.',
			path: '../../../../docs/legal/PRIVACY_POLICY.md',
		},
		en: {
			version: '1.0',
			updated: '2026-08-31',
			title: 'Privacy Policy for Itinera',
			description:
				'Information on the processing of personal data in Itinera in accordance with the GDPR and Spanish data protection law.',
			path: '../../../../docs/legal/PRIVACY_POLICY_EN.md',
		},
	},
} as const;

export const SUPPORTED_LOCALES = ['es', 'en'] as const;
export type LegalLocale = (typeof SUPPORTED_LOCALES)[number];