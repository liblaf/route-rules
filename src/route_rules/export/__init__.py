# tangerine-start: lazy-loader.py
import lazy_loader as _lazy

__getattr__, __dir__, __all__ = _lazy.attach_stub(__name__, __file__)
# tangerine-end
