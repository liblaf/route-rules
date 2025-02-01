import route_rules.container.optim._utils as u


def merge_domain_with_suffix(
    domain: set[str], domain_suffix: set[str]
) -> tuple[set[str], set[str]]:
    domain_new: set[str] = set()
    suffix_new: set[str] = domain_suffix.copy()
    for d in domain:
        if "." + d in suffix_new:
            suffix_new.remove("." + d)
            suffix_new.add(d)
            continue
        if u.match_domain_suffix(d, suffix_new):
            continue
        domain_new.add(d)
    return domain_new, suffix_new
