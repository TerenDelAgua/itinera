# RAT — Registro de Actividades de Tratamiento

> **Documento interno, no público. Requisito Art. 30 RGPD.**
> **Versión:** 1.1
> **Última actualización:** 12 de agosto de 2026
> **Responsable:** Juan Carlos Del Agua Pascual (TEREN)

---

## Metadatos del responsable

| Campo | Detalle |
|-------|---------|
| Responsable | Juan Carlos Del Agua Pascual (TEREN) |
| NIF | 48411328P |
| Dirección | Calle Llíria 49, puerta 5, 46185 La Eliana, Valencia, España |
| Email de contacto | privacy@teren.dev |
| Persona de contacto RGPD | Juan Carlos Del Agua Pascual (el titular) — privacy@teren.dev |
| DPO | No designado (no obligatorio para esta actividad) |
| Delegado ante AEPD | No aplica |
| Fuente de los datos | Directamente del interesado (formulario de registro, navegación en la app) |
| Medidas organizativas | Acceso limitado al responsable único (autónomo). Procedimiento documentado de respuesta a incidentes (notificación AEPD en 72h según Art. 33 RGPD). Revisión anual del RAT. Política de privacidad y términos de servicio públicos y vigentes. |

---

## Actividad 1: Gestión de usuarios y cuentas

| Campo | Detalle |
|-------|---------|
| Finalidad | Prestación del servicio de planificación de viajes |
| Base legal | Art. 6.1.b RGPD (ejecución de contrato) para email, hash de contraseña, fecha de registro, idioma y tier. Art. 6.1.f RGPD (interés legítimo en seguridad) para IP y User-Agent. |
| Interesados | Usuarios registrados e invitados |
| Datos | Email, hash de contraseña, IP, User-Agent, fecha de registro, idioma, tier |
| Encargados del tratamiento | Railway (hosting DB), Vercel (hosting frontend), Resend (emails) |
| Destinatarios | Ninguno fuera de los encargados |
| Transferencias internacionales | Sí, a EE.UU. Bajo el Data Privacy Framework UE-EE.UU. (Decisión (UE) 2023/1795) y, subsidiariamente, las Cláusulas Contractuales Tipo (Decisión (UE) 2021/914). |
| Plazo de conservación | Cuenta activa + 30 días post-eliminación |
| Medidas técnicas | Bcrypt cost 12, tokens opacos rotativos, rate limiting, HTTPS, cookies HttpOnly/Secure/SameSite, cifrado en reposo AES-256 |

---

## Actividad 2: Envío de comunicaciones transaccionales

| Campo | Detalle |
|-------|---------|
| Finalidad | Emails de bienvenida y recuperación de contraseña |
| Base legal | Art. 6.1.b RGPD (ejecución de contrato) |
| Interesados | Usuarios registrados |
| Datos | Email, idioma |
| Encargados del tratamiento | Resend (envío de emails) |
| Destinatarios | Ninguno fuera de los encargados |
| Transferencias internacionales | Sí, a EE.UU. Bajo el Data Privacy Framework UE-EE.UU. (Decisión (UE) 2023/1795) y, subsidiariamente, las Cláusulas Contractuales Tipo (Decisión (UE) 2021/914). |
| Plazo de conservación | No se almacenan; envío puntual |
| Medidas técnicas | DPA con Resend, tokens de un solo uso para reset |

---

## Actividad 3: Seguridad y prevención de abuso

| Campo | Detalle |
|-------|---------|
| Finalidad | Rate limiting, detección de fuerza bruta, logging técnico |
| Base legal | Art. 6.1.f RGPD (interés legítimo) |
| Interesados | Todos los visitantes (guests y registrados) |
| Datos | IP, User-Agent, timestamp, endpoint accedido |
| Encargados del tratamiento | Railway (almacenamiento de logs) |
| Destinatarios | Ninguno (uso interno) |
| Transferencias internacionales | No (logs almacenados en infraestructura UE de Railway) |
| Plazo de conservación | Máximo 30 días |
| Medidas técnicas | Rate limiting DB-backed, eliminación automática por expiración |

---

## Proceso de gestión de derechos RGPD

Canal para ejercer derechos: privacy@teren.dev o desde la app (Configuración → Eliminar cuenta para supresión).

Plazo de respuesta: máximo 1 mes desde la solicitud, prorrogable a 2 meses adicionales en casos complejos (Art. 12.3 RGPD).

Verificación de identidad: podemos solicitar información adicional para verificar la identidad del solicitante (Art. 12.6 RGPD).

Derechos soportados: acceso, rectificación, supresión, limitación, portabilidad, oposición, no ser objeto de decisiones automatizadas.

Registro interno: cada solicitud se registra con fecha de recepción, fecha de respuesta y acciones tomadas (responsabilidad proactiva, Art. 5.2 RGPD).

---

## Proceso de notificación de brechas de seguridad

En caso de violación de seguridad que afecte a datos personales:

1. Notificación a la AEPD en un plazo máximo de 72 horas desde que tengamos constancia (Art. 33.1 RGPD).
2. Notificación a los usuarios afectados sin dilación indebida cuando la brecha sea susceptible de afectarles de forma significativa (Art. 34.1 RGPD).
3. Documentación interna de la brecha con: naturaleza, categorías y número de afectados, consecuencias probables, medidas adoptadas (Art. 33.5 RGPD).

---

## Revisión de este documento

Este RAT se revisará al menos una vez al año y siempre que se produzcan cambios materiales en los tratamientos descritos (nuevas actividades, nuevos encargados, cambios de finalidad, etc.).

---

## Firma

**Juan Carlos Del Agua Pascual**
Fecha: 12 de agosto de 2026

---

**Changelog:**

- **v1.1** (12 ago 2026) — Base legal diferenciada por tipo de dato en Actividad 1 (Art. 6.1.b para email/hash, Art. 6.1.f para IP/UA). Eliminado "nombre de usuario (si se añade)" de Actividad 2 (campo inexistente, no especulación). Añadida Persona de contacto RGPD en metadatos. Distinguidos "Encargados del tratamiento" vs "Destinatarios" en las 3 actividades. Añadidas referencias jurídicas específicas para transferencias internacionales (Decisión DPF UE-EE.UU. 2023/1795 y SCC Decisión 2021/914). Añadidas medidas organizativas. Añadida sección de proceso de gestión de derechos RGPD (Art. 12.3, 12.6, 15-22). Añadida sección de proceso de notificación de brechas (Art. 33-34). Añadida periodicidad de revisión. Añadida fuente de los datos. Eliminado campo redundante "Categorías especiales" de Actividad 1. Añadido cifrado en reposo AES-256 a medidas técnicas de Actividad 1.
- **v1.0** (12 ago 2026) — Versión inicial.
