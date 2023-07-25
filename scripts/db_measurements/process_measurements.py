def convert_to_MB(s):
    s = s.replace(',', '.')
    n = float(s[:-1])
    t = s[-1]
    if t == 'K':
        return str(int(n / 1024))
    if t == 'M':
        return str(int(n))
    if t == 'G':
        return str(int(n * 1024))
    return "oops"


with open('measurements.txt', 'r') as file:
    ans = []
    for line in file:
        ln = line.strip().split()
        ind = ln.index("data:")
        ans.append(ln[ind + 1])
    print(' '.join(map(convert_to_MB, ans)))
        