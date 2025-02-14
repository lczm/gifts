const API_URL = import.meta.env.VITE_API_URL;

export const lookupTeam = async (staffPassId: string) => {
  const response = await fetch(
    `${API_URL}/lookup?staff_pass_id=${staffPassId}`
  );
  if (!response.ok) {
    throw new Error("Network response was not ok");
  }
  return response.json();
};

export const redeemGift = async (staffPassId: string) => {
  const response = await fetch(`${API_URL}/redemption`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ staff_pass_id: staffPassId }),
  });
  const data = await response.json();
  return data;
};
