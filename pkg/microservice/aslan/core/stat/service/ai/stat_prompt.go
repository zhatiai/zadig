package ai

const ParseUserPromptPrompt = `
你是文本信息提取器，需要严格根据下面 1 2这两条规则来提取输入 {} 内的内容。
文本提取器规则为下列1. 2. 三条规则：
1. 项目列表：返回三重引号分割的输入数据中的项目名称列表。
举例1：\"\"\"请通过构建，部署这些工作流任务来分析 min-test、cosmos-helm-1 项目三个月以来的质量是否有提升\"\"\"'，此时匹配到的项目列表是：{"min-test","cosmos-helm-1"}。
举例2：\"\"\"分析一下最近几天项目的发展趋势\"\"\",则项目列表返回空数组。
举例3：\"\"\"分析项目min-test，tt-zhati未来三个月的趋势\"\"\",则匹配到的项目列表为["min-test","tt-zhati"]。
2. job列表：job列表是指项目的构建(build)、测试(test)、部署(deploy)、发布(release)这四个过程，请在三重引号分割的输入数据中匹配这四个过程，如果匹配到，则返回匹配到的job列表，
如果没有匹配到，则返回空数组；在匹配完后，你的结果必须是这个数组的子集[\"build\",\"test\",\"deploy\",\"release\"]。
举例1：输入数据：\"\"\"请通过构建，部署这些工作流任务来分析 min-test、cosmos-helm-1 项目三个月以来的质量是否有提升\"\"\"，此刻匹配到的job列表是:["build","deploy"]
举例2：输入数据：\"\"\"请分析项目min-test最近两个月的情况\"\"\",此时job列表返回空 "job_list":[]

下面()中的内容是用户和文本提取器的交互过程的案例：
(输入：{请通过构建，测试这些工作流任务来分析 helm-test-2、cosmos-helm-1 项目三个月以来的质量是否有提升}
文本信息提取器返回结果：{"project_list":["helm-test-2", "cosmos-helm-1"],"job_list":["build", "test"]}")。
输入: {请分析项目min-test最近两个月的情况}
文本信息提取器返回结果: {"project_list":["min-test"],"job_list":[]})

请按前面的规则提取下面 三重引号 内的内容，同时需要校验自己的答案是否符合提取规则和输出要求。输出要求：仅返回JSON数据，不要有任何多余解释过程的输出。
`

const EveryProjectAnalysisPrompt = `首先，你需要按照以下两个步骤来完成回答：第一步：你需要根据项目数据计算每周的构建成功和失败率，部署成功和失败率，测试成功和失败率，必须要对计算结果做检查，
判断是否正确，否则重新计算；第二步：你需要根据项目数据分析该项目的三个指标：
1. 项目质量：质量由测试成功率和测试覆盖率决定，各指标成功率越高，质量越好，注意，如果该项目的测试成功率和失败率为0，则代表该项目质量不可靠，需要引起重视；
2. 项目效率：效率由构建、部署、测试的成功率以及构建，部署，测试的平均耗时决定，成功率越高，平均耗时越低，效率越高，注意，如果该项目的构建、部署、测试的成功率和失败率同时为0，则代表该项目异常，需要引起重视；
3. 缺陷：缺陷由测试、构建、部署的失败率来决定，失败率越高，可能存在问题的概率越大，同时测试覆盖率低，存在问题概率越大。
在你的回答中需要包含第一步计算的的每周的构建、部署、测试的成功率和失败率和第二步对三个指标的分析结果，同时也要包含对这个项目存在问题的分析并给出高质量建议，回答格式要满足三个大括号的回答示例：
{{{项目名称：xxx，数据范围是2023年4月1号到2023年6月1号，2023年4月1号-2023年4月8号，构建次数为，部署次数....;1. 项目质量：该项目总共执行了35次测试，其中测试成功次数为30，失败次数为5次；
2. 项目效率：该项目构建了40次，其中构建成功了30次，失败了10次,部署....;3. 该项目存在以下问题....;通过对项目数据的分析，我认为该项目测试次数过少....需要注意...可以通过增加测试次数来提高项目质量....}}}
你的回答不要超过400个汉字。`

const ProjectAnalysisPrompt = `假设你是资深Devops专家，你需要根据分析要求去分析三重引号分割的项目数据，分析要求：%s;你需要根据所有的项目数据并根据分析要求进行多角度分析；如果存在多个项目，你需要在分析时候从构建、测试、部署、发布几个角度来进行对比并深度分析；
你的回答需要使用text格式输出, 输出内容不要包含\"三重引号分割的项目数据\"这个名称，也不要复述分析要求中的内容,需要给出明确的分析结论,在你的回答中禁止包含 \"data_description\" 字段; 项目数据：\"\"\"%s\"\"\";`

// PromptExamples great example to display
var PromptExamples = []string{
	"通过历史数据，请用简洁的文字总结%s项目最近一个月的整体表现。",
	"请根据项目%s的构建、部署、测试和发布等数据，分析项目最近一个月的现状，并基于历史数据，分析未来的趋势和潜在问题，并提出改进建议。",
	"根据%s项目最近一个月每周的构建，部署，测试等数据的变化，以此分析该项目最近一段时间的发展趋势，如果存在问题则分析原因并给出合理的解决方案。",
	"通过历史数据，分析%s项目的最大短板是什么？并针对这些短板提供一些解决办法。",
	"分析所有项目%s在最近一个月的整体表现，选出质量和效率最高的一个项目和最差的一个项目,并分析这两个项目产生差距的原因。",
	"从项目质量和效率两个角度分析项目%s最近一个月的情况，并分析这些项目的最近一个月的构建和部署趋势，对比构建趋势分析这些项目发展情况。",
}

const AttentionPrompt = `你需要从两个月数据中做对比，分析出两个月中变化最大的5条变化的内容，输出格式为json格式，你可以根据三个括号分割的样例来输出一样格式的内容，内容来源于你的分析结果,样例:(((
{
  \"answer\": [
    {
      \"project\": \"项目A\",
      \"result\": \"项目A测试成功率提升了47.2%\",
      \"name\": \"test_success_rate\",
      \"current_month\": \"77.20\",
      \"last_month\": \"30.00\"
    },
    {
      \"project\": \"项目B\",
      \"result\": \"项目B需求交付周期延长5天\",
      \"name\": \"requirement_development_lead_time\",
      \"current_month\": \"20\",
      \"last_month\": \"15\"
    },
    {
      \"project\": \"项目R\",
      \"result\": \"项目R发布频次提升20%\",
      \"name\": \"release_frequency\",
      \"current_month\": \"60.20\",
      \"last_month\": \"40.20\"
    }
  ]
})));
你可以按照以下步骤来进行分析：
1.步骤一，计算current_month和last_month的构建成功率，测试成功率，部署成功率，发布成功率，发布频次，以及需求研发周期等数据的差值，你必须对计算结果做检查，判断是否正确，否则重新计算；
2.步骤二，根据差值的绝对值从大到小排序，取前三个差值最大的数据；
3.步骤三，根据差值的正负，分别输出三条变化的内容，如果差值为正，则输出“项目A测试成功率提升了40%”，如果差值为负，则输出“项目A测试成功率下降了40%”。
你的输出必须是JSON格式，返回前必须对你的回答进行检查，并根据两个月差异的从大到小进行排序再返回，保证差异最大的内容在最上面，有错误则重新进行计算并分析，如果项目数大于5，则返回变化最大的5条内容，如果项目数量小于5条，则返回变化最大的3条内容，如果没有数据，则返回如下JSON内容{\n    \"answer\":[]\n}。
`
