# Política de Privacidad — /privacy

**Versión ES (gobernante)**
**Política de Privacidad de Itinera**
**Última actualización:** 31 de agosto de 2026 · **Versión:** 1.2

---

## 1. Responsable del tratamiento

| Campo | Dato |
|-------|------|
| Servicio | Itinera (proyecto de teren.dev) |
| Email de contacto | privacy@teren.dev |
| Web del titular | https://teren.dev |

Itinera es un proyecto personal operado bajo el dominio teren.dev. No se ha designado Delegado de Protección de Datos (DPO) al no ser obligatorio para esta actividad. Para cualquier consulta sobre privacidad, contacta en el email indicado.

## 2. Qué datos tratamos

Itinera es una aplicación de planificación de viajes. Tratamos los siguientes datos personales:

| Categoría | Datos concretos | Cuándo |
|-----------|-----------------|--------|
| Identificativos | Email | Al crear cuenta |
| Autenticación | Hash de contraseña (bcrypt) | Al crear cuenta |
| Técnicos | Dirección IP, User-Agent | En cada request (seguridad y rate limiting) |
| Contenido | Viajes, itinerarios, gastos, notas | Al usar la app |
| Metadatos de cuenta | Fecha de registro, idioma preferido, tier | Al crear cuenta |

Lo que NO hacemos:
- No usamos cookies de tracking ni analytics de terceros.
- No compartimos datos con anunciantes.
- No vendemos datos. Nunca.
- No usamos fingerprinting ni identificadores de publicidad.

## 3. Finalidades y base legal

| Finalidad | Base legal (RGPD) |
|-----------|-------------------|
| Prestación del servicio (crear, editar, guardar viajes) | Art. 6.1.b — Ejecución de contrato |
| Gestión de cuenta (registro, login, recuperación de contraseña) | Art. 6.1.b — Ejecución de contrato |
| Envío de emails transaccionales (bienvenida, reset de contraseña) | Art. 6.1.b — Ejecución de contrato |
| Seguridad (rate limiting, prevención de abuso) | Art. 6.1.f — Interés legítimo |
| Cumplimiento de obligaciones legales | Art. 6.1.c — Obligación legal |

No realizamos ningún tratamiento basado en consentimiento. Todos los tratamientos se basan en ejecución de contrato o interés legítimo.

## 4. Destinatarios y encargados del tratamiento

No cedemos datos a terceros salvo obligación legal. Utilizamos los siguientes proveedores como encargados de tratamiento:

| Encargado | Servicio | Ubicación | Medidas |
|-----------|----------|-----------|---------|
| Railway (Railway Corp.) | Hosting del backend y base de datos | EE.UU. / UE | DPA estándar de Railway |
| Vercel (Vercel Inc.) | Hosting del frontend | Global (CDN) | DPA estándar de Vercel |
| Resend (Resend Inc.) | Envío de emails transaccionales | EE.UU. | DPA estándar de Resend |

Para transferencias fuera del EEE, nos basamos en las Cláusulas Contractuales Tipo de la Comisión Europea (SCC) y/o el Data Privacy Framework UE-EE.UU.

## 5. Conservación de datos

| Dato | Plazo de conservación |
|------|----------------------|
| Cuenta activa | Mientras la cuenta exista |
| Tras solicitud de eliminación | 30 días (soft delete) + anonimización permanente |
| Viajes de usuarios eliminados | Anonimizados (se elimina la vinculación) |
| Sesiones inactivas | 30 días (expiración automática) |
| Logs técnicos (IP, User-Agent) | Máximo 30 días |

## 6. Tus derechos (Art. 15-22 RGPD)

Tienes derecho a:
- **Acceso** — Obtener copia de tus datos.
- **Rectificación** — Corregir datos inexactos.
- **Supresión** — Eliminar tu cuenta y todos tus datos.
- **Limitación** — Restringir el tratamiento temporalmente.
- **Portabilidad** — Recibir tus datos en formato estructurado (JSON).
- **Oposición** — Oponerte al tratamiento basado en interés legítimo.
- **No ser objeto de decisiones automatizadas con efectos jurídicos** — No aplicamos decisiones basadas únicamente en tratamiento automatizado que produzcan efectos sobre ti (Art. 22 RGPD). Las medidas de seguridad automatizadas (rate limiting, detección de abuso) no producen efectos jurídicos.

**Verificación de identidad:** podemos solicitar información adicional para verificar tu identidad antes de atender la solicitud (Art. 12.6 RGPD). Esto es para proteger tu cuenta.

Cómo ejercerlos:
- Desde la app: Configuración → Eliminar cuenta (supresión inmediata).
- Por email: Envía tu solicitud a privacy@teren.dev indicando qué derecho deseas ejercer.

**Plazos:** respondemos en un máximo de 1 mes desde la recepción de la solicitud, prorrogable a 2 meses adicionales en casos complejos, notificándote la prórroga y sus motivos dentro del primer mes (Art. 12.3 RGPD). El ejercicio es gratuito.

**Reclamaciones:** si consideras que no hemos atendido correctamente tu solicitud, puedes reclamar ante la Agencia Española de Protección de Datos (AEPD): www.aepd.es

## 7. Seguridad

Implementamos medidas técnicas y organizativas adecuadas:
- Contraseñas hasheadas con bcrypt (cost 12).
- Sesiones con tokens opacos rotativos (no JWT).
- Rate limiting contra fuerza bruta.
- Conexiones cifradas (HTTPS/TLS).
- Cifrado en reposo: AES-256 en los volúmenes de Base de datos.
- Cookies HttpOnly, Secure, SameSite.
- Sin exposición de tokens en URL o logs.

En caso de violación de seguridad que afecte a tus datos, te notificaremos sin dilación indebida y a la AEPD en un plazo máximo de 72 horas.

## 8. Cookies

Itinera solo utiliza cookies estrictamente necesarias para el funcionamiento del servicio:

| Cookie | Finalidad | Datos que almacena | Duración |
|--------|-----------|---------------------|----------|
| `session_id` | Identificar sesión de invitado | UUID v4 (no contiene datos personales) | 1 año |
| `itinera_access` | Autenticación de usuario | Token opaco aleatorio (32 bytes). En el servidor se almacena hasheado, nunca en plano. | 24 horas |
| `itinera_refresh` | Renovación de sesión | Token opaco aleatorio (32 bytes, rotativo). En el servidor se almacena hasheado, nunca en plano. | 30 días |

No usamos cookies de terceros, analytics, publicidad ni fingerprinting. Al ser cookies estrictamente necesarias, no requieren consentimiento previo (Art. 22.2 LSSI).

## 9. Menores

Itinera no está dirigida a menores de 14 años. No recopilamos deliberadamente datos de menores. Si detectamos que un menor ha creado una cuenta, la eliminaremos.

## 10. Cambios en esta política

Si realizamos cambios materiales, te avisaremos con antelación razonable (mínimo 14 días) mediante notificación en la app o email. La versión vigente es la publicada en esta página con su fecha de última actualización.

Las versiones anteriores quedan archivadas en `/privacy/archive` y pueden consultarse en cualquier momento.

## 11. Contacto

Para cualquier consulta sobre privacidad: privacy@teren.dev

## 12. Ley aplicable y jurisdicción

Las presentes condiciones se rigen por la legislación española y la normativa europea de protección de datos aplicable. Para cualquier controversia derivada de esta política, las partes se someten a los juzgados y tribunales de Valencia, España, sin perjuicio de los derechos que te asisten como consumidor ante tu domicilio habitual en el EEE.

---
