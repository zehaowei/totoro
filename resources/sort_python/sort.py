import random
import time


def quick_sort(arr, s, e):
    if s >= e:
        return
    candi = arr[s]
    i = s
    j = s+1
    while j <= e:
        if arr[j] < candi:
            if i != j-1:
                tmp = arr[i+1]
                arr[i+1] = arr[j]
                arr[j] = tmp
            i += 1
        j += 1
    arr[s] = arr[i]
    arr[i] = candi
    quick_sort(arr, s, i-1)
    quick_sort(arr, i+1, e)


start = time.time()
array = []
for i in range(int(1e7)):
    array.append(random.randint(1, 1000000))
    if i % 1000000 == 0:
        print(int(time.time() - start))
quick_sort(array, 0, int(1e7)-1)

print("finish sort: ", int(time.time() - start))
