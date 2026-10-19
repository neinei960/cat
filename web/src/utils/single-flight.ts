export function singleFlight<TArgs extends unknown[], TResult>(
  action: (...args: TArgs) => Promise<TResult>,
) {
  let inFlight = false

  return async (...args: TArgs): Promise<TResult | undefined> => {
    if (inFlight) return undefined
    inFlight = true
    try {
      return await action(...args)
    } finally {
      inFlight = false
    }
  }
}
