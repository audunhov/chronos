export type ChronosFetcherExtraProps = {
  /**
   * You can add some extra props to your generated fetchers.
   *
   * Note: You need to re-gen after adding the first property to
   * have the `ChronosFetcherExtraProps` injected in `ChronosComponents.ts`
   **/
};

const baseUrl = "";

export type ErrorWrapper<TError> =
  | TError
  | { status: "unknown"; payload: string };

export type ChronosFetcherOptions<TBody, THeaders, TQueryParams, TPathParams> =
  {
    url: string;
    method: string;
    body?: TBody;
    headers?: THeaders;
    queryParams?: TQueryParams;
    pathParams?: TPathParams;
    signal?: AbortSignal;
  } & ChronosFetcherExtraProps;

export async function chronosFetch<
  TData,
  TError,
  TBody extends {} | FormData | undefined | null,
  THeaders extends {},
  TQueryParams extends {},
  TPathParams extends {},
>({
  url,
  method,
  body,
  headers,
  pathParams,
  queryParams,
  signal,
}: ChronosFetcherOptions<
  TBody,
  THeaders,
  TQueryParams,
  TPathParams
>): Promise<TData> {
  let error: ErrorWrapper<TError>;
  try {
    const token = window.localStorage.getItem('chronos_auth') 
      ? JSON.parse(window.localStorage.getItem('chronos_auth')!).token 
      : '';

    const requestHeaders: HeadersInit = {
      "Content-Type": "application/json",
      ...headers,
    };

    if (token) {
      (requestHeaders as any)["Authorization"] = `Bearer ${token}`;
    }

    /**
     * As the fetch API is being used, when multipart/form-data is specified
     * the Content-Type header must be deleted so that the browser can set
     * the correct boundary.
     * https://developer.mozilla.org/en-US/docs/Web/API/FormData/Using_FormData_Objects#sending_files_using_a_formdata_object
     */
    if (
      requestHeaders["Content-Type"]
        ?.toLowerCase()
        .includes("multipart/form-data")
    ) {
      delete requestHeaders["Content-Type"];
    }

    const response = await window.fetch(
      `${baseUrl}${resolveUrl(url, queryParams, pathParams)}`,
      {
        signal,
        method: method.toUpperCase(),
        body: body
          ? body instanceof FormData
            ? body
            : JSON.stringify(body)
          : undefined,
        headers: requestHeaders,
      },
    );
    if (!response.ok) {
      try {
        error = await response.json();
      } catch (e) {
        error = {
          status: "unknown" as const,
          payload:
            e instanceof Error
              ? `Unexpected error (${e.message})`
              : "Unexpected error",
        };
      }
    } else if (response.headers.get("content-type")?.includes("json")) {
      return await response.json();
    } else {
      // if it is not a json response, assume it is a blob and cast it to TData
      return (await response.blob()) as unknown as TData;
    }
  } catch (e) {
    const errorObject: Error = {
      name: "unknown" as const,
      message:
        e instanceof Error ? `Network error (${e.message})` : "Network error",
      stack: e as string,
    };
    throw errorObject;
  }
  throw error;
}

const resolveUrl = (
  url: string,
  queryParams: Record<string, string | undefined | null> = {},
  pathParams: Record<string, string> = {},
) => {
  const cleanQueryParams = Object.fromEntries(
    Object.entries(queryParams).filter(([_, value]) => value !== undefined && value !== null)
  ) as Record<string, string>;

  let query = new URLSearchParams(cleanQueryParams).toString();
  if (query) query = `?${query}`;
  return (
    url.replace(/\{\w*\}/g, (key) => pathParams[key.slice(1, -1)] ?? "") + query
  );
};
