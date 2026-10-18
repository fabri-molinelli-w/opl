export const redirectToLogin = (): void => {
  if (typeof window !== 'undefined') {
    window.location.assign('/login')
  }
}
