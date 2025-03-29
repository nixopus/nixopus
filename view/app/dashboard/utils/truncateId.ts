const truncateId = (id: string) => {
  return id?.substring(0, 12) || '';
};

export default truncateId;
