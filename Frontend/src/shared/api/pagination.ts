export type PagedResult<T> = {
  items: T[];
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
};

export type PageParams = {
  page?: number;
  perPage?: number;
  search?: string;
};

export function pageQuery(params: PageParams = {}) {
  const query = new URLSearchParams();
  query.set("page", String(params.page ?? 1));
  query.set("per_page", String(params.perPage ?? 10));
  if (params.search?.trim()) query.set("search", params.search.trim());
  return `?${query.toString()}`;
}
