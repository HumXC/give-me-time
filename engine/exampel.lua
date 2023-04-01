-- E 是 Element 的缩写
-- E 是所有在 json 中定义的

-- 按下一个元素
press(E.main.start)
-- 按下某个坐标
press(12, 222)
-- 滑动，用 300 毫秒从 start 滑动到坐标 (12, 21)
swipe(E.main.start).to(12, 21).action(300)
-- ocr 识别元素内透明部分的内容
ocr(E.main.uid)
-- ocr 识别一片区域的内容，ocr 不会使用 Offset
ocr(0, 0, 12, 323)
-- 查找某个元素的坐标，这个坐标已经计算过 Offset
-- 这个元素必须有 img 字段
x, y, v = find(E.game.button)

-- 当调用的函数有 element 参数，或者有需要图像识别的场景时，
-- 例如 press(E.main.start) 会先从设备截图，然后通过图像识别找到这个 element 的位置，
-- 根据 json 中的定义再根据某些规则计算出点击的位置再进行按下操作。
-- 而 press(12, 233) 这种调用方式不需要截图，会立马发送按下操作，因为直接提供了点击的坐标。
-- 每一次调用需要图像识别的函数时都会重新截图，而有些时候需要识别同一张截图，
-- 例如在主界面中，我需要获取“背包”和“开始”按钮的位置，那么可以使用 find() 函数，
-- 这两个按钮在同一个页面下，并且 find() 函数也不会改变界面，所以只需要截一张图就行
-- 而如果直接写:
p = find(E.main.pacbag)
s = find(E.main.start)
-- 此时会进行两次截图操作，为了省下一次截图操作，可以使用 lock() 和 unlock() 函数
lock()
p = find(E.main.pacbag)
s = find(E.main.start)
unlock()
-- 在调用 lock() 时会对屏幕进行截图，直到调用 unlock() 前，所有需要图像识别的函数都会使用同一张截图而不是每次都重新截图
-- lock() 和 unlock() 总是成对出现，无法连续 lock() 或者 unlock()

-- 获取屏幕截图
-- s 是一个 table
s = screen()
-- 设置 “元素” 操作所要处理的图片
lock(s)
-- 以下代码会在截图中查找 E.main 元素并点击
s = screen()
lock(s)
press(E.main)
unlock()
-- screen() 函数获取的图片的数据会被暂时保存在内存中，需要手动释放
s1 = screen()
lock(s1)
unlock()

s1.free() -- 释放这张截图的内存，之后 s1 将不再可用
lock(s1) -- 报错，因为 s1 已经被释放了
-- 将截图写入一个文件
s1.save("./screen.jpg")

-- M 是 Mechine 的缩写
-- 执行 adb 命令
M.adb()