-- E 是 Element 的缩写
-- E 是所有在 json 中定义的
-- 按下一个元素 100 毫秒
press(E.main.start, 100)
-- 按下某个坐标
press(12, 222)
-- 滑动，用 300 毫秒从 start 滑动到坐标 (12, 21) 再用 100 毫秒滑动到 title
swipe(E.main.start).to(12, 21, 300).to(E.main.title, 100).do()
-- ocr 识别元素内透明部分的内容
ocr(E.main.uid)
-- ocr 识别一片区域的内容
ocr(0, 0, 12, 323)
-- 查找某个元素的左上角坐标
find(E.game.button)
-- M 是 Mechine 的缩写
-- 执行 adb 命令
M.adb()