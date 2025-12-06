from concurrent.futures import ThreadPoolExecutor, as_completed
from typing import Callable, List, Optional, TypeVar

T = TypeVar("T")
R = TypeVar("R")


def process_parallel(
    items: List[T],
    processor_func: Callable[[T], R],
    max_workers: int = 50,
    error_handler: Optional[Callable[[T, Exception], R]] = None,
) -> List[R]:
    """Process items in parallel using ThreadPoolExecutor"""
    if not items:
        return []

    results = []
    max_workers = min(len(items), max_workers)

    with ThreadPoolExecutor(max_workers=max_workers) as executor:
        futures = {executor.submit(processor_func, item): item for item in items}

        for future in as_completed(futures):
            try:
                result = future.result()
                results.append(result)
            except Exception as e:
                item = futures[future]
                if error_handler:
                    error_result = error_handler(item, e)
                    results.append(error_result)
    return results

