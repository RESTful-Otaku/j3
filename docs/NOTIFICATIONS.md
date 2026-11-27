# Notifications

In-app `notify*` + `PAToast` + `ToastHost`, `j3-toast`, `notifySuccess(message, undo)`, `runUndo`. `POST /api/v1/demo/announcements` records + `state.changed` broadcast + SSE `announcement`.

**Fan-out** `fanoutNotifications` inserts `notification_deliveries` per recipient x channel for kinds: `announcement, info, feedback, confirmation, account, preference, invite, marketing`. Channels `push,email,sms`. Audience `all, fighters/participants, officials, clubs, guardians, spectators, admins`. Recipients from `users`. Local delivery always. Optional `J3_NOTIFY_WEBHOOK_URL` posts JSON per delivery; failures recorded `failed`. Without credentials `recorded` + `local`. No fake "sent" for external providers. Migration `0026_notification_deliveries.sql`.
