const getStatusColor = (status: string) => {
  if (status?.toLowerCase().includes('running')) return 'bg-green-100 text-green-800 rounded-full';
  if (status?.toLowerCase().includes('exited')) return 'bg-red-100 text-red-800 rounded-full';
  if (status?.toLowerCase().includes('created')) return 'bg-blue-100 text-blue-800 rounded-full';
  return 'bg-gray-100 text-gray-800 rounded-full';
};

export default getStatusColor;
