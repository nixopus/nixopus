export const parsePort = (port: string) => {
  const parsedPort = parseInt(port, 10);
  return isNaN(parsedPort) ? null : parsedPort;
};
