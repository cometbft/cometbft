function hasOwn(obj, key) {
  return Object.prototype.hasOwnProperty.call(obj, key);
}

// Escape special characters.
function escapeRe(str) {
  return str.replace(/[.*+?^$|[\](){}\\-]/g, '\\$&');
}

// Return a future date by the given string.
function computeExpires(str) {
  const lastCh = str.charAt(str.length - 1);
  const value = parseInt(str, 10);
  let expires = new Date();

  switch (lastCh) {
    case 'Y': expires.setFullYear(expires.getFullYear() + value); break;
    case 'M': expires.setMonth(expires.getMonth() + value); break;
    case 'D': expires.setDate(expires.getDate() + value); break;
    case 'h': expires.setHours(expires.getHours() + value); break;
    case 'm': expires.setMinutes(expires.getMinutes() + value); break;
    case 's': expires.setSeconds(expires.getSeconds() + value); break;
    default: expires = new Date(str);
  }

  return expires;
}

// Convert an object to a cookie option string.
function convert(opts) {
  let res = '';

  // eslint-disable-next-line
  for (const key in opts) {
    if (hasOwn(opts, key)) {
      if (/^expires$/i.test(key)) {
        let expires = opts[key];

        if (typeof expires !== 'object') {
          expires += typeof expires === 'number' ? 'D' : '';
          expires = computeExpires(expires);
        }
        res += `;${key}=${expires.toUTCString()}`;
      } else if (/^secure$/.test(key)) {
        if (opts[key]) {
          res += `;${key}`;
        }
      } else {
        res += `;${key}=${opts[key]}`;
      }
    }
  }

  if (!hasOwn(opts, 'path')) {
    res += ';path=/';
  }

  return res;
}

export {
  hasOwn,
  escapeRe,
  computeExpires,
  convert,
};
