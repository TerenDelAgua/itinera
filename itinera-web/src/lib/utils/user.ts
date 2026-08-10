/**
 * User-presentation helpers.
 *
 * `getInitials` derives the avatar label
 * from the email's local part. The contract is fixed:
 *   - `juan.carlos@x.com` → `JU`
 *   - `a@x.com`            → `A` (only one char available)
 *   - `..@x.com`           → `?` (empty after strip → fallback)
 *   - `''`                 → `?` (defensive: never render an empty
 *                            avatar circle)
 *
 * The strip removes `_`, `.`, `-` because those are common separators
 * in email local-parts and produce noise in the avatar (`._`).
 */
export function getInitials(email: string | null | undefined): string {
    if (!email) return '?';
    const at = email.indexOf('@');
    const local = at >= 0 ? email.slice(0, at) : email;
    const cleaned = local.replace(/[._-]/g, '').toUpperCase();
    if (cleaned.length === 0) return '?';
    if (cleaned.length === 1) return cleaned;
    return cleaned.slice(0, 2);
}