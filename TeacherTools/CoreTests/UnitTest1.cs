using Core;

namespace CoreTests
{
    [TestClass]
    public class UnitTest1
    {
        [TestMethod]
        public void TestMethod1()
        {
            var class1 = new Class1();

            var result = class1.Add(1, 1);

            Assert.AreEqual(2, result);
        }
    }
}