-- 测试 "E"
-- E 是在 test.json 定义的 Element 在 lua 中的根节点。
-- E 是一个嵌套的 table，其子元素也是 table。
-- 每一个元素都必须有一个 "_name" 字段用“点分”的形式表示其层级关系
-- 例如 E.main.start 的 _name 就是 "main.start" -> E.main.start["_name"]="main.start"
function assertString(a, b) 
    if type(a) ~= "string" then
        error("assert failed [" .. a .. "] is not a string")
    end
    if type(b) ~= "string" then
        error("assert failed [" .. b .. "] is not a string")
    end
    if a ~= b then
        error("assert failed: [" .. a .."] is not equal [" .. b .. "]")
    end
end

assertString(E.main._name, "main")
assertString(E.main.start._name, "main.start")
assertString(E.main.text._name, "main.text")
assertString(E.main.text.input._name, "main.text.input")
assertString(E.game._name, "game")

-- 测试全局函数
-- 调用这个函数说明 fn 必须发生错误
function assertError(fnName, fn, ...)
    local noterr = pcall(fn,...)
    if noterr then 
        error("function [name:" .. fnName .. "] should be ab error, but not")
    end
end
-- 合法调用
press(0, 5)
press(67, 35, 0)
press(1, 3, 300)
press(E.main.start, 100)
-- 非法调用
assertError("press", press, E)
assertError("press", press, "main")
assertError("press", press, E.main.start, -120)