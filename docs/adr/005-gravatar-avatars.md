# ADR 005: Gravatar Avatars

## Context

The MVP needs user avatars on public profile pages without introducing file uploads, object storage, image resizing or moderation.

## Decision

Use Gravatar URLs derived from the user's email address.

## Consequences

Positive:
- No file storage required
- Simple implementation
- Works for public profiles and dashboard

Negative:
- Avatar depends on email
- Users cannot upload custom images yet
