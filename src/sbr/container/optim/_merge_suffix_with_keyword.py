import sbr.container.optim._utils as u


def merge_suffix_with_keyword(
    domain_suffix: set[str], domain_keyword: set[str]
) -> tuple[set[str], set[str]]:
    suffix_new: set[str] = set()
    keyword_new: set[str] = domain_keyword.copy()
    for s in domain_suffix:
        if u.match_domain_keyword(s, keyword_new):
            continue
        suffix_new.add(s)
    return suffix_new, keyword_new
