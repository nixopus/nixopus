const ENVIRONMENT_REGEX = /^[a-z0-9]+(-[a-z0-9]+)*$/;

export function formatEnvironmentName(input: string): string {
  return input
    .toLowerCase()
    .replace(/[_\s]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

export function isValidEnvironmentName(value: string): boolean {
  return value.length >= 1 && value.length <= 50 && ENVIRONMENT_REGEX.test(value);
}
