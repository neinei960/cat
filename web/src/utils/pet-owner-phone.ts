// Three-way merge: a remote contact edit must not discard a local draft.
export function reconcileOwnerPhone(baseline: string, draft: string, remote: string) {
  const clean = draft.trim() === baseline.trim()
  return {
    phone: clean ? remote : draft,
    conflict: !clean && remote.trim() !== baseline.trim() && remote.trim() !== draft.trim(),
  }
}
