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