type Decoder<T = string> = (value: string) => T;
type Encoder<T = string> = (value: T) => string;

interface Cookies<T = string> {
  [key: string] : T
}

export interface CookieOptions {
  domain?: string
  path?: string
  expires?: string | number
  'max-age'?: number
  secure?: boolean
  samesite?: string
}

declare function isEnabled() : boolean;

/**
 * Get the cookie value by key.
 */
declare function get<T = string>(key: string, decoder?: Decoder<T>) : T | null;

/**
 * Get all cookies
 */
declare function getAll<T = string>(decoder?: Decoder<T>) : Cookies<T>;

/**
 * Set a cookie.
 */
declare function set<T = string>(key: string, value: T, encoder : Encoder<T>, options?: CookieOptions) : void;

/**
 * Set a cookie.
 */
declare function set(key: string, value: string, options?: CookieOptions) : void;

/**
 * Remove a cookie by the specified key.
 */
declare function remove(key: string, options?: CookieOptions) : void;

/**
 * Get the cookie's value without decoding.
 */
declare function getRaw(key: string) : string | null;

/**
 * Set a cookie without encoding the value.
 */
declare function setRaw(key: string, value: string, options?: CookieOptions) : void;

export {
  isEnabled,
  get,
  getAll,
  set,
  getRaw,
  setRaw,
  remove,
  isEnabled as isCookieEnabled,
  get as getCookie,
  getAll as getAllCookies,
  set as setCookie,
  getRaw as getRawCookie,
  setRaw as setRawCookie,
  remove as removeCookie
}
