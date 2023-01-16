(function (global, factory) {
  typeof exports === 'object' && typeof module !== 'undefined' ? factory(exports) :
  typeof define === 'function' && define.amd ? define(['exports'], factory) :
  (global = global || self, factory(global.Cookie = {}));
}(this, function (exports) { 'use strict';

  function _extends() {
    _extends = Object.assign || function (target) {
      for (var i = 1; i < arguments.length; i++) {
        var source = arguments[i];

        for (var key in source) {
          if (Object.prototype.hasOwnProperty.call(source, key)) {
            target[key] = source[key];
          }
        }
      }

      return target;
    };

    return _extends.apply(this, arguments);
  }

  function hasOwn(obj, key) {
    return Object.prototype.hasOwnProperty.call(obj, key);
  } // Escape special characters.


  function escapeRe(str) {
    return str.replace(/[.*+?^$|[\](){}\\-]/g, '\\$&');
  } // Return a future date by the given string.


  function computeExpires(str) {
    var lastCh = str.charAt(str.length - 1);
    var value = parseInt(str, 10);
    var expires = new Date();

    switch (lastCh) {
      case 'Y':
        expires.setFullYear(expires.getFullYear() + value);
        break;

      case 'M':
        expires.setMonth(expires.getMonth() + value);
        break;

      case 'D':
        expires.setDate(expires.getDate() + value);
        break;

      case 'h':
        expires.setHours(expires.getHours() + value);
        break;

      case 'm':
        expires.setMinutes(expires.getMinutes() + value);
        break;

      case 's':
        expires.setSeconds(expires.getSeconds() + value);
        break;

      default:
        expires = new Date(str);
    }

    return expires;
  } // Convert an object to a cookie option string.


  function convert(opts) {
    var res = ''; // eslint-disable-next-line

    for (var key in opts) {
      if (hasOwn(opts, key)) {
        if (/^expires$/i.test(key)) {
          var expires = opts[key];

          if (typeof expires !== 'object') {
            expires += typeof expires === 'number' ? 'D' : '';
            expires = computeExpires(expires);
          }

          res += ";" + key + "=" + expires.toUTCString();
        } else if (/^secure$/.test(key)) {
          if (opts[key]) {
            res += ";" + key;
          }
        } else {
          res += ";" + key + "=" + opts[key];
        }
      }
    }

    if (!hasOwn(opts, 'path')) {
      res += ';path=/';
    }

    return res;
  }

  function isEnabled() {
    var key = '@key@';
    var value = '1';
    var re = new RegExp("(?:^|; )" + key + "=" + value + "(?:;|$)");
    document.cookie = key + "=" + value + ";path=/";
    var enabled = re.test(document.cookie);

    if (enabled) {
      // eslint-disable-next-line
      remove(key);
    }

    return enabled;
  } // Get the cookie value by key.


  function get(key, decoder) {
    if (decoder === void 0) {
      decoder = decodeURIComponent;
    }

    if (typeof key !== 'string' || !key) {
      return null;
    }

    var reKey = new RegExp("(?:^|; )" + escapeRe(key) + "(?:=([^;]*))?(?:;|$)");
    var match = reKey.exec(document.cookie);

    if (match === null) {
      return null;
    }

    return typeof decoder === 'function' ? decoder(match[1]) : match[1];
  } // The all cookies


  function getAll(decoder) {
    if (decoder === void 0) {
      decoder = decodeURIComponent;
    }

    var reKey = /(?:^|; )([^=]+?)(?:=([^;]*))?(?:;|$)/g;
    var cookies = {};
    var match;
    /* eslint-disable no-cond-assign */

    while (match = reKey.exec(document.cookie)) {
      reKey.lastIndex = match.index + match.length - 1;
      cookies[match[1]] = typeof decoder === 'function' ? decoder(match[2]) : match[2];
    }

    return cookies;
  } // Set a cookie.


  function set$1(key, value, encoder, options) {
    if (encoder === void 0) {
      encoder = encodeURIComponent;
    }

    if (typeof encoder === 'object' && encoder !== null) {
      /* eslint-disable no-param-reassign */
      options = encoder;
      encoder = encodeURIComponent;
      /* eslint-enable no-param-reassign */
    }

    var attrsStr = convert(options || {});
    var valueStr = typeof encoder === 'function' ? encoder(value) : value;
    var newCookie = key + "=" + valueStr + attrsStr;
    document.cookie = newCookie;
  } // Remove a cookie by the specified key.


  function remove(key, options) {
    var opts = {
      expires: -1
    };

    if (options) {
      opts = _extends({}, options, opts);
    }

    return set$1(key, 'a', opts);
  } // Get the cookie's value without decoding.


  function getRaw(key) {
    return get(key, null);
  } // Set a cookie without encoding the value.


  function setRaw(key, value, options) {
    return set$1(key, value, null, options);
  }

  exports.isEnabled = isEnabled;
  exports.isCookieEnabled = isEnabled;
  exports.get = get;
  exports.getCookie = get;
  exports.getAll = getAll;
  exports.getAllCookies = getAll;
  exports.set = set$1;
  exports.setCookie = set$1;
  exports.getRaw = getRaw;
  exports.getRawCookie = getRaw;
  exports.setRaw = setRaw;
  exports.setRawCookie = setRaw;
  exports.remove = remove;
  exports.removeCookie = remove;

  Object.defineProperty(exports, '__esModule', { value: true });

}));
