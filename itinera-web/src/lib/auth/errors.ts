/**
 * Frontend mapping of the backend's error code registry.

 * The messages here are *fallback* copies of the ones the backend
 * emits in its 4 supported locales. They are used ONLY when the
 * server message cannot be shown (e.g. network errors, generic
 * catch-all branches, edge cases where the backend returned an
 * unknown code).
 */

export type ErrorCode =
    | 'EMAIL_ALREADY_EXISTS'
    | 'EMAIL_INVALID'
    | 'PASSWORD_TOO_WEAK'
    | 'TERMS_NOT_ACCEPTED'
    | 'VALIDATION_ERROR'
    | 'INVALID_CREDENTIALS'
    | 'ACCOUNT_DELETED'
    | 'RATE_LIMITED'
    | 'UNAUTHENTICATED'
    | 'INTERNAL_ERROR'
    | 'NETWORK_ERROR'
    | 'UNKNOWN_ERROR';

export type Locale = 'en' | 'es' | 'ja' | 'id';

interface ErrorMessages {
    en: string;
    es: string;
    ja: string;
    id: string;
}

const MESSAGES: Record<ErrorCode, ErrorMessages> = {
    EMAIL_ALREADY_EXISTS: {
        en: 'This email is already in use. Sign in instead?',
        es: 'Este email ya está en uso. ¿Iniciamos sesión?',
        ja: 'このメールアドレスは既に使われています。ログインしますか？',
        id: 'Email ini sudah digunakan. Masuk saja?'
    },
    EMAIL_INVALID: {
        en: 'That email address looks invalid.',
        es: 'Ese email no parece válido.',
        ja: 'メールアドレスの形式が正しくありません。',
        id: 'Format email tidak valid.'
    },
    PASSWORD_TOO_WEAK: {
        en: 'Use at least 8 characters with a number or symbol.',
        es: 'Usa al menos 8 caracteres con un número o símbolo.',
        ja: '8文字以上で、数字または記号を含めてください。',
        id: 'Gunakan minimal 8 karakter dengan angka atau simbol.'
    },
    TERMS_NOT_ACCEPTED: {
        en: 'Please accept the terms to continue.',
        es: 'Acepta los términos para continuar.',
        ja: '続行するには利用規約に同意してください。',
        id: 'Harap setujui syarat untuk melanjutkan.'
    },
    VALIDATION_ERROR: {
        en: 'Please review the highlighted fields.',
        es: 'Revisa los campos marcados.',
        ja: 'ハイライト表示された項目を確認してください。',
        id: 'Periksa kolom yang ditandai.'
    },
    INVALID_CREDENTIALS: {
        // Anti-enumeration: the same string is returned whether the email
        // does not exist or the password is wrong. Spec 017 §5.2 + §9.2.
        en: 'Incorrect email or password.',
        es: 'Email o contraseña incorrectos.',
        ja: 'メールアドレスまたはパスワードが正しくありません。',
        id: 'Email atau kata sandi salah.'
    },
    ACCOUNT_DELETED: {
        // Soft-deleted accounts cannot sign in. The user is told to
        // re-register after the 30-day retention window (Spec §5.9) so
        // the same email becomes available again with a fresh hash.
        en: 'This account was deleted. You can register again after 30 days.',
        es: 'Esta cuenta fue eliminada. Puedes registrarte de nuevo en 30 días.',
        ja: 'このアカウントは削除されました。30日後に再登録できます。',
        id: 'Akun ini dihapus. Anda dapat mendaftar lagi setelah 30 hari.'
    },
    RATE_LIMITED: {
        en: 'Too many attempts. Please wait a moment.',
        es: 'Demasiados intentos. Espera un momento.',
        ja: '試行回数が多すぎます。しばらくお待ちください。',
        id: 'Terlalu banyak percobaan. Tunggu sebentar.'
    },
    UNAUTHENTICATED: {
        en: 'You need to sign in to continue.',
        es: 'Inicia sesión para continuar.',
        ja: '続行するにはログインしてください。',
        id: 'Masuk untuk melanjutkan.'
    },
    INTERNAL_ERROR: {
        en: 'Something went wrong. Please try again.',
        es: 'Algo ha ido mal. Inténtalo de nuevo.',
        ja: '問題が発生しました。もう一度お試しください。',
        id: 'Terjadi kesalahan. Silakan coba lagi.'
    },
    NETWORK_ERROR: {
        en: 'No connection. Check your network and try again.',
        es: 'Sin conexión. Comprueba tu red e inténtalo de nuevo.',
        ja: '接続できません。ネットワークを確認してください。',
        id: 'Tidak ada koneksi. Periksa jaringan Anda dan coba lagi.'
    },
    UNKNOWN_ERROR: {
        en: 'Something went wrong. Please try again.',
        es: 'Algo ha ido mal. Inténtalo de nuevo.',
        ja: '問題が発生しました。もう一度お試しください。',
        id: 'Terjadi kesalahan. Silakan coba lagi.'
    }
};

/**
 * Resolve the localised message for a given error code.
 *
 * @param code     machine identifier (e.g. `EMAIL_ALREADY_EXISTS`)
 * @param locale   one of the supported locales; falls back to `en` if
 *                 the requested locale is missing a translation
 * @param fallback returned when `code` is not in the registry — usually
 *                 the server's own localised message (preferred over
 *                 guessing)
 */
export function resolveErrorMessage(
    code: string,
    locale: Locale,
    fallback?: string
): string {
    const entry = (MESSAGES as Record<string, ErrorMessages | undefined>)[code];
    if (!entry) {
        // Unknown code: prefer the caller's fallback (which is the
        // server's own localised message, more authoritative than ours),
        // then fall back to the generic UNKNOWN_ERROR message in the
        // requested locale.
        return fallback ?? MESSAGES.UNKNOWN_ERROR[locale] ?? MESSAGES.UNKNOWN_ERROR.en;
    }
    return entry[locale] ?? entry.en;
}

/**
 * Type guard that narrows a string to a known ErrorCode. Returns
 * `UNKNOWN_ERROR` for unknown codes so callers always have a safe
 * fallback.
 */
export function toErrorCode(code: string): ErrorCode {
    return (MESSAGES as Record<string, ErrorMessages | undefined>)[code]
        ? (code as ErrorCode)
        : 'UNKNOWN_ERROR';
}
